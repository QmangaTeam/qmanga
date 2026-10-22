#!/usr/bin/env python3
"""Upload prepared manga pages through Qmanga API. Python 3.10+, no dependencies."""
import argparse
import hashlib
import ipaddress
import json
import os
from pathlib import Path, PurePosixPath
import re
import stat
import sys
import tempfile
import time
import urllib.error
import urllib.parse
import urllib.request
import zipfile

MAX_PAGE_BYTES = 25 * 1024 * 1024
MAX_TOTAL_BYTES = 500 * 1024 * 1024
EXTENSIONS = {'.jpg', '.jpeg', '.png', '.webp'}

class UploadError(Exception):
    pass

class APIError(UploadError):
    def __init__(self, status, code):
        super().__init__(f'API returned {status} ({code})')
        self.status = status


def validate_base(value):
    parsed = urllib.parse.urlsplit(value)
    try:
        local = parsed.hostname == 'localhost' or ipaddress.ip_address(parsed.hostname or '').is_loopback
    except ValueError:
        local = False
    if parsed.username or parsed.password or parsed.query or parsed.fragment or parsed.path not in ('', '/') or not parsed.hostname or (parsed.scheme != 'https' and not (parsed.scheme == 'http' and local)):
        raise UploadError('Use an HTTPS origin, or HTTP on localhost, without credentials, path or query.')
    return value.rstrip('/')


def read_token(path):
    if path.is_symlink() or not path.is_file() or (os.name == 'posix' and path.stat().st_mode & 0o077):
        raise UploadError('Token file must be a regular private file (chmod 600).')
    token = path.read_text().strip()
    if not token or any(c.isspace() for c in token):
        raise UploadError('Invalid token file.')
    return token


def natural_key(value):
    return tuple((1, int(x)) if x.isdigit() else (0, x.casefold()) for x in re.split(r'(\d+)', value))


def collect_pages(source):
    """Inventory and hash all input before contacting API; never extract ZIP paths."""
    result = []
    if source.is_symlink():
        raise UploadError('Source cannot be a symlink.')
    if source.is_dir():
        entries = sorted((p for p in source.iterdir() if p.suffix.lower() in EXTENSIONS), key=lambda p: (natural_key(p.name), p.name))
        for p in entries:
            if p.is_symlink() or not p.is_file():
                raise UploadError('Page must be a regular file.')
            if p.stat().st_size > MAX_PAGE_BYTES:
                raise UploadError('Page exceeds 25 MiB.')
        readers = [(p.name, p.stat().st_size, lambda p=p: p.read_bytes()) for p in entries]
        consume_pages(readers, result)
    elif source.suffix.lower() in ('.cbz', '.zip') and source.is_file():
        with zipfile.ZipFile(source) as archive:
            entries = archive.infolist()
            if len(entries) > 5000:
                raise UploadError('Archive contains too many entries.')
            seen = set()
            for entry in entries:
                name = entry.filename
                path = PurePosixPath(name)
                if name in seen or '\\' in name or path.is_absolute() or '..' in path.parts or re.match(r'^[A-Za-z]:', name) or stat.S_ISLNK(entry.external_attr >> 16) or entry.flag_bits & 1:
                    raise UploadError('Unsafe, duplicate or encrypted archive entry.')
                seen.add(name)
            images = sorted((e for e in entries if not e.is_dir() and PurePosixPath(e.filename).suffix.lower() in EXTENSIONS), key=lambda e: (natural_key(e.filename), e.filename))
            consume_pages([(e.filename, e.file_size, lambda e=e: archive.read(e)) for e in images], result)
    else:
        raise UploadError('Source must be an image folder or CBZ/ZIP archive.')
    if not result:
        raise UploadError('No JPEG, PNG or WebP pages found.')
    return result


def consume_pages(readers, result):
    if not 1 <= len(readers) <= 300 or sum(size for _, size, _ in readers) > MAX_TOTAL_BYTES:
        raise UploadError('Chapter limit: 1–300 pages, 500 MiB total.')
    for position, (name, size, read) in enumerate(readers, 1):
        if not 0 < size <= MAX_PAGE_BYTES:
            raise UploadError('Invalid page size (maximum 25 MiB).')
        data = read()
        if len(data) != size:
            raise UploadError('Source changed while reading.')
        digest = hashlib.sha256(data).hexdigest()
        result.append({'name': name, 'position': position, 'size': size, 'sha256': digest, 'key': f'cli-v1-{position}-{digest}'})


def page_bytes(source, page):
    if source.is_dir():
        path = source / page['name']
        if path.is_symlink() or path.stat().st_size != page['size']:
            raise UploadError('Source changed after inventory.')
        data = path.read_bytes()
    else:
        with zipfile.ZipFile(source) as archive:
            entry = archive.getinfo(page['name'])
            if entry.file_size != page['size']:
                raise UploadError('Archive changed after inventory.')
            data = archive.read(entry)
    if len(data) != page['size'] or hashlib.sha256(data).hexdigest() != page['sha256']:
        raise UploadError('Page content changed; use a new import manifest/state.')
    return data


def save_state(path, state):
    if path.is_symlink():
        raise UploadError('State path cannot be a symlink.')
    fd, temporary = tempfile.mkstemp(prefix='.qmanga-', dir=path.parent)
    try:
        with os.fdopen(fd, 'w') as out:
            json.dump(state, out, ensure_ascii=False, indent=2)
            out.flush(); os.fsync(out.fileno())
        os.replace(temporary, path)
    finally:
        if os.path.exists(temporary): os.unlink(temporary)


def load_state(path, fingerprint, base, actor):
    if not path.exists():
        return {'fingerprint': fingerprint, 'base': base, 'actor': actor}
    if path.is_symlink(): raise UploadError('State cannot be a symlink.')
    state = json.loads(path.read_text())
    if any(state.get(k) != v for k, v in [('fingerprint', fingerprint), ('base', base), ('actor', actor)]):
        raise UploadError('State belongs to a different input, site or account; do not reuse it.')
    return state


class NoRedirect(urllib.request.HTTPRedirectHandler):
    def redirect_request(self, req, fp, code, msg, headers, newurl):
        raise UploadError('API redirect refused to protect the bearer token.')


class API:
    def __init__(self, base, token):
        self.base = validate_base(base)
        self.token = token
        self.opener = urllib.request.build_opener(NoRedirect())
    def call(self, method, path, body=None, raw=False):
        if not path.startswith('/api/v1/') or path.startswith('//'):
            raise UploadError('Invalid API path.')
        data = body if raw else (json.dumps(body).encode() if body is not None else None)
        request = urllib.request.Request(self.base + path, data=data, method=method, headers={'Authorization': 'Bearer ' + self.token, 'Content-Type': 'application/octet-stream' if raw else 'application/json'})
        try:
            with self.opener.open(request, timeout=120) as response:
                data = response.read(8 * 1024 * 1024 + 1)
                if len(data) > 8 * 1024 * 1024: raise UploadError('API response too large.')
                return json.loads(data) if data else None
        except urllib.error.HTTPError as failure:
            # Response bodies may contain supplied text; report only a validated machine code.
            try: code = json.loads(failure.read(4096)).get('error', {}).get('code', 'request_failed')
            except (ValueError, AttributeError): code = 'request_failed'
            if not isinstance(code, str) or not re.fullmatch(r'[a-z_]{1,60}', code): code = 'request_failed'
            raise APIError(failure.code, code) from None
        except urllib.error.URLError:
            raise UploadError('Connection failed. Rerun the same command with the same state file.') from None


def all_items(api, path):
    for offset in range(0, 1000001, 100):
        batch = api.call('GET', path + ('&' if '?' in path else '?') + 'offset=' + str(offset))['items']
        yield from batch
        if len(batch) < 100: return
    raise UploadError('Pagination limit exceeded.')


def validate_draft(revision, pages):
    expected = {p['key']: p for p in pages}
    for page in revision.get('pages', []):
        wanted = expected.get(page.get('upload_key'))
        if not wanted or page.get('position') != wanted['position']:
            raise UploadError('Draft contains other pages/order. Import stopped without deleting or replacing them.')


def prepare_target(api, manifest, state, save):
    if state.get('project_id'):
        project = api.call('GET', '/api/v1/studio/projects/' + quote(state['project_id']))
        if state.get('title_id') and state['title_id'] != project['title_id']: raise UploadError('Project/title mismatch.')
    elif manifest.get('project_id'):
        project = api.call('GET', '/api/v1/studio/projects/' + quote(manifest['project_id']))
        state.update(project_id=project['id'], title_id=project['title_id']); save()
    else:
        title_id = state.get('title_id') or manifest.get('title_id')
        if not title_id:
            title = manifest['title']
            match = next((t for t in all_items(api, '/api/v1/admin/titles?q=' + quote(title['slug'])) if t['slug'] == title['slug']), None)
            if match: title_id = match['id']
            else:
                try: title_id = api.call('POST', '/api/v1/admin/titles', title)['id']
                except APIError as error:
                    if error.status != 409: raise
                    match = next((t for t in all_items(api, '/api/v1/admin/titles?q=' + quote(title['slug'])) if t['slug'] == title['slug']), None)
                    if not match: raise
                    title_id = match['id']
            state['title_id'] = title_id; save()
        detail = api.call('GET', '/api/v1/admin/titles/' + quote(title_id))
        team = manifest['team_id']
        match = next((p for p in detail['projects'] if p['team_id'] == team), None)
        if match: project_id = match['id']
        else:
            try: project_id = api.call('POST', '/api/v1/admin/titles/' + quote(title_id) + '/projects', {'team_id': team})['id']
            except APIError as error:
                if error.status != 409: raise
                detail = api.call('GET', '/api/v1/admin/titles/' + quote(title_id))
                match = next((p for p in detail['projects'] if p['team_id'] == team), None)
                if not match: raise
                project_id = match['id']
        state.update(title_id=title_id, project_id=project_id); save()
    if not state.get('chapter_id'):
        if manifest.get('chapter_id'): state['chapter_id'] = manifest['chapter_id']
        else:
            chapter = manifest['chapter']
            def find():
                return next((c for c in all_items(api, '/api/v1/studio/projects/' + quote(state['project_id']) + '/chapters') if c['order'] == chapter['order']), None)
            match = find()
            if not match:
                try: match = api.call('POST', '/api/v1/admin/titles/' + quote(state['title_id']) + '/chapters', chapter)
                except APIError as error:
                    if error.status != 409: raise
                    match = find()
                    if not match: raise
            if 'label' in match and (match['label'] != chapter['label'] or match.get('name', '') != chapter.get('name', '')):
                raise UploadError('Existing chapter at this order has a different label/name. Specify chapter_id explicitly.')
            state['chapter_id'] = match['id']
        save()


def quote(value):
    return urllib.parse.quote(str(value), safe='')


def run_upload(api, manifest, source, pages, state_path, publish, submit, timeout, poll=2):
    actor = api.call('GET', '/api/v1/auth/me')['user']['id']
    fingerprint = hashlib.sha256(json.dumps({'manifest': manifest, 'pages': pages}, sort_keys=True, ensure_ascii=False).encode()).hexdigest()
    state = load_state(state_path, fingerprint, api.base, actor)
    save = lambda: save_state(state_path, state)
    save(); prepare_target(api, manifest, state, save)
    if not state.get('revision_id'):
        rev = api.call('POST', '/api/v1/studio/projects/' + quote(state['project_id']) + '/revisions', {'chapter_id': state['chapter_id']})
        state['revision_id'] = rev['id']; state['release_id'] = rev['release_id']; save()
    path = '/api/v1/studio/revisions/' + quote(state['revision_id'])
    revision = api.call('GET', path)
    if revision.get('project_id') != state['project_id'] or revision.get('chapter_id') != state['chapter_id']: raise UploadError('Revision target mismatch.')
    validate_draft(revision, pages)
    if revision['state'] == 'published':
        if len(revision['pages']) != len(pages) or revision.get('release_status') != 'published' or revision.get('current_revision_id') != revision['id']: raise UploadError('Stored revision is no longer the active publication.')
        return {'state': 'published', 'release_id': revision['release_id'], 'revision_id': revision['id']}
    if revision['state'] == 'submitted' and submit and len(revision['pages']) == len(pages):
        return {'state': 'submitted', 'release_id': revision['release_id'], 'revision_id': revision['id']}
    if revision['state'] not in ('draft', 'processing', 'ready'): raise UploadError('Revision is immutable or in moderation; no files changed.')
    existing = {p['upload_key']: p for p in revision['pages']}
    for page in pages:
        old = existing.get(page['key'])
        if old and old['status'] in ('ready', 'queued', 'processing'): continue
        if old and old['status'] == 'failed' and old.get('failure') == 'processing_failed':
            api.call('POST', path + '/pages/' + quote(old['id']) + '/retry', {})
        else:
            query = urllib.parse.urlencode({'key': page['key'], 'position': page['position'], 'name': PurePosixPath(page['name']).name})
            api.call('PUT', path + '/pages?' + query, page_bytes(source, page), raw=True)
        print(f"Uploaded {page['position']}/{len(pages)}", file=sys.stderr)
    deadline = time.monotonic() + timeout
    while True:
        revision = api.call('GET', path); validate_draft(revision, pages)
        statuses = [p['status'] for p in revision['pages']]
        if len(statuses) == len(pages) and all(s == 'ready' for s in statuses): break
        if any(s == 'failed' for s in statuses): raise UploadError('Page processing failed; inspect revision, then rerun to retry retained sources.')
        if time.monotonic() >= deadline: raise UploadError('Processing wait timed out. Rerun the same command to resume.')
        time.sleep(poll)
    attribution = {k: manifest[k] for k in ('source_note', 'credits', 'rights_confirmed', 'preparation_method', 'human_reviewed')}
    api.call('PUT', path + '/attribution', {'version': revision['version'], **attribution})
    cover = manifest.get('cover_page')
    if cover:
        selected = next(p for p in revision['pages'] if p['upload_key'] == pages[cover - 1]['key'])
        api.call('PUT', '/api/v1/studio/projects/' + quote(state['project_id']) + '/cover', {'page_id': selected['id']})
    revision = api.call('GET', path)
    if publish or submit:
        api.call('POST', path + ('/publish' if publish else '/submit'), {'version': revision['version']})
        revision = api.call('GET', path)
        expected = 'published' if publish else 'submitted'
        if revision['state'] != expected: raise UploadError('Server did not confirm requested publication state.')
    state['result'] = revision['state']; save()
    return {'state': revision['state'], 'release_id': revision['release_id'], 'revision_id': revision['id']}


def validate_manifest(manifest, count):
    if not isinstance(manifest, dict): raise UploadError('Manifest must be a JSON object.')
    if not manifest.get('project_id') and not (manifest.get('team_id') and (manifest.get('title_id') or isinstance(manifest.get('title'), dict))):
        raise UploadError('Specify project_id, or team_id and title/title_id.')
    if not manifest.get('chapter_id') and not isinstance(manifest.get('chapter'), dict): raise UploadError('Specify chapter_id or chapter metadata.')
    if not manifest.get('chapter_id'):
        c = manifest['chapter']
        if not isinstance(c.get('label'), str) or not c['label'].strip() or type(c.get('order')) is not int or not 1 <= c['order'] <= 1000000:
            raise UploadError('Chapter needs label and positive integer order.')
    if manifest.get('rights_confirmed') is not True or not isinstance(manifest.get('source_note'), str) or len(manifest['source_note'].strip()) < 8 or not isinstance(manifest.get('credits'), str) or len(manifest['credits'].strip()) < 2:
        raise UploadError('Explicit rights_confirmed:true, source_note (8+ chars) and credits (2+ chars) are required.')
    if manifest.get('preparation_method') not in ('unspecified', 'manual', 'machine_assisted', 'machine') or type(manifest.get('human_reviewed')) is not bool:
        raise UploadError('Specify preparation_method and truthful human_reviewed boolean.')
    cover = manifest.get('cover_page')
    if cover is not None and (type(cover) is not int or not 1 <= cover <= count): raise UploadError('cover_page must be a 1-based input page index.')


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--base-url', required=True)
    parser.add_argument('--manifest', type=Path, required=True)
    parser.add_argument('--source', type=Path, required=True)
    parser.add_argument('--state', type=Path, help='Keep this file to resume; default MANIFEST.qmanga-state.json')
    parser.add_argument('--token-file', type=Path, help='Private file; otherwise QMANGA_API_TOKEN environment variable')
    group = parser.add_mutually_exclusive_group(); group.add_argument('--publish', action='store_true', help='Publish directly with token permission'); group.add_argument('--submit', action='store_true', help='Send to normal moderation')
    parser.add_argument('--timeout', type=int, default=600, help='Processing wait seconds (default600)')
    args = parser.parse_args()
    try:
        token = read_token(args.token_file) if args.token_file else os.environ.get('QMANGA_API_TOKEN', '').strip()
        if not token or any(c.isspace() for c in token): raise UploadError('Set QMANGA_API_TOKEN or use --token-file. Never pass a token on the command line.')
        api = API(args.base_url, token)
        if args.timeout < 1: raise UploadError('timeout must be positive.')
        manifest = json.loads(args.manifest.read_text())
        pages = collect_pages(args.source); validate_manifest(manifest, len(pages))
        state = args.state or args.manifest.with_suffix('.qmanga-state.json')
        result = run_upload(api, manifest, args.source, pages, state, args.publish, args.submit, args.timeout)
        print(json.dumps(result, ensure_ascii=False))
        return 0
    except (UploadError, OSError, ValueError, KeyError, TypeError, zipfile.BadZipFile):
        # Avoid tracebacks that might expose HTTP headers or private metadata.
        error = sys.exc_info()[1]
        print(str(error) if isinstance(error, UploadError) else 'Invalid input or unreadable file. Check manifest/source/state.', file=sys.stderr)
        return 1

if __name__ == '__main__':
    raise SystemExit(main())
