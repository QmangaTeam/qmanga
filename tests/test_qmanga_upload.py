import importlib.util
from pathlib import Path
import tempfile
import unittest
import unittest.mock
import zipfile
import json
import os

SPEC = importlib.util.spec_from_file_location('qmanga_upload', Path(__file__).parents[1] / 'scripts/qmanga-upload.py')
cli = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(cli)

class ImportTests(unittest.TestCase):
    def test_folder_natural_order_and_repeated_identical_pages_have_distinct_keys(self):
        with tempfile.TemporaryDirectory() as folder:
            for name in ['10.png', '2.png', '1.png']:
                Path(folder, name).write_bytes(b'prepared image')
            pages = cli.collect_pages(Path(folder))
            self.assertEqual([p['name'] for p in pages], ['1.png', '2.png', '10.png'])
            self.assertEqual(len({p['key'] for p in pages}), 3)
            self.assertEqual(pages, cli.collect_pages(Path(folder)))

    def test_zip_rejects_traversal_even_in_non_image_entry(self):
        with tempfile.TemporaryDirectory() as folder:
            path = Path(folder, 'bad.cbz')
            with zipfile.ZipFile(path, 'w') as archive:
                archive.writestr('001.png', b'page')
                archive.writestr('../escape.txt', b'bad')
            with self.assertRaises(cli.UploadError):
                cli.collect_pages(path)

    def test_zip_rejects_duplicate_names_and_symlinks(self):
        with tempfile.TemporaryDirectory() as folder:
            path = Path(folder, 'bad.cbz')
            with zipfile.ZipFile(path, 'w') as archive:
                item = zipfile.ZipInfo('001.png'); item.external_attr = 0o120777 << 16
                archive.writestr(item, 'elsewhere')
            with self.assertRaises(cli.UploadError): cli.collect_pages(path)

    def test_quota_and_symlink_are_checked_before_http(self):
        with tempfile.TemporaryDirectory() as folder:
            path = Path(folder, '001.png'); path.write_bytes(b'image')
            os.symlink(path, Path(folder, '002.png'))
            with self.assertRaises(cli.UploadError): cli.collect_pages(Path(folder))
            Path(folder, '002.png').unlink()
            with unittest.mock.patch.object(cli, 'MAX_PAGE_BYTES', 2):
                with self.assertRaises(cli.UploadError): cli.collect_pages(Path(folder))

    def test_token_file_requires_private_permissions(self):
        with tempfile.TemporaryDirectory() as folder:
            path = Path(folder, 'token'); path.write_text('qmanga_test_secret'); path.chmod(0o644)
            with self.assertRaises(cli.UploadError): cli.read_token(path)
            path.chmod(0o600)
            self.assertEqual(cli.read_token(path), 'qmanga_test_secret')

    def test_base_url_refuses_remote_plaintext_credentials_or_fragments(self):
        for bad in ['http://example.org', 'https://owner:secret@example.org', 'https://example.org/#token', 'https://example.org/?token=bad']:
            with self.assertRaises(cli.UploadError): cli.validate_base(bad)
        self.assertEqual(cli.validate_base('http://localhost:18080/'), 'http://localhost:18080')

    def test_resume_refuses_changed_input_or_account(self):
        with tempfile.TemporaryDirectory() as folder:
            path = Path(folder, 'state.json'); cli.save_state(path, {'fingerprint':'a', 'base':'https://example.org','actor':'owner'})
            self.assertEqual(cli.load_state(path,'a','https://example.org','owner')['actor'],'owner')
            for args in [('b','https://example.org','owner'), ('a','https://different.org','owner'), ('a','https://example.org','other')]:
                with self.assertRaises(cli.UploadError): cli.load_state(path,*args)
            self.assertEqual(path.stat().st_mode & 0o777,0o600)

    def test_lost_publish_response_resumes_without_new_draft_or_duplicate_upload(self):
        with tempfile.TemporaryDirectory() as folder:
            source = Path(folder, 'pages'); source.mkdir(); (source / '001.png').write_bytes(b'prepared')
            pages = cli.collect_pages(source)
            manifest = {'project_id':'p','chapter_id':'c','source_note':'Test synthetic source','credits':'Test','rights_confirmed':True,'preparation_method':'machine','human_reviewed':False}
            class FakeAPI:
                base = 'https://example.org'
                drafts = uploads = publications = 0
                revision = {'id':'r','project_id':'p','chapter_id':'c','release_id':'release','state':'draft','version':1,'pages':[]}
                def call(self, method, path, body=None, raw=False):
                    if path == '/api/v1/auth/me': return {'user':{'id':'owner'}}
                    if path == '/api/v1/studio/projects/p': return {'id':'p','title_id':'t'}
                    if path.endswith('/revisions'):
                        self.drafts += 1; return self.revision
                    if '/pages?' in path:
                        self.uploads += 1
                        self.revision.update(state='ready',version=2,pages=[{'id':'page','upload_key':pages[0]['key'],'position':1,'status':'ready'}]); return {}
                    if path.endswith('/attribution'):
                        self.revision['version'] += 1; return None
                    if path.endswith('/publish'):
                        self.publications += 1
                        self.revision.update(state='published',release_status='published',current_revision_id='r')
                        raise cli.UploadError('Simulated lost response')
                    return self.revision
            api = FakeAPI(); state = Path(folder, 'state.json')
            with self.assertRaises(cli.UploadError): cli.run_upload(api,manifest,source,pages,state,True,False,1,poll=0)
            result = cli.run_upload(api,manifest,source,pages,state,True,False,1,poll=0)
            self.assertEqual(result['state'],'published')
            self.assertEqual((api.drafts,api.uploads,api.publications),(1,1,1))
            api.revision['current_revision_id']='newer-revision'
            with self.assertRaises(cli.UploadError): cli.run_upload(api,manifest,source,pages,state,True,False,1,poll=0)

    def test_lost_submission_response_reports_existing_submission(self):
        with tempfile.TemporaryDirectory() as folder:
            source = Path(folder, 'pages'); source.mkdir(); (source/'001.png').write_bytes(b'image')
            pages = cli.collect_pages(source)
            manifest = {'project_id':'p','chapter_id':'c'}
            fingerprint = cli.hashlib.sha256(json.dumps({'manifest':manifest,'pages':pages},sort_keys=True,ensure_ascii=False).encode()).hexdigest()
            state = Path(folder,'state.json'); cli.save_state(state,{'fingerprint':fingerprint,'base':'https://example.org','actor':'owner','project_id':'p','chapter_id':'c','revision_id':'r'})
            class FakeAPI:
                base='https://example.org'
                def call(self,method,path,body=None,raw=False):
                    if method!='GET': raise AssertionError('Resumed immutable submission must not mutate')
                    if path.endswith('/auth/me'): return {'user':{'id':'owner'}}
                    if path.endswith('/projects/p'): return {'id':'p','title_id':'t'}
                    return {'id':'r','release_id':'release','project_id':'p','chapter_id':'c','state':'submitted','pages':[{'upload_key':pages[0]['key'],'position':1}]}
            self.assertEqual(cli.run_upload(FakeAPI(),manifest,source,pages,state,False,True,1)['state'],'submitted')

    def test_concurrent_chapter_conflict_cannot_adopt_different_name(self):
        class FakeAPI:
            calls=0
            def call(self,method,path,body=None,raw=False):
                if path.endswith('/projects/p'): return {'id':'p','title_id':'t'}
                if '/chapters?' in path:
                    self.calls+=1
                    return {'items':[] if self.calls==1 else [{'id':'other','label':'2','name':'Other chapter','order':1}]}
                raise cli.APIError(409,'conflict')
        with self.assertRaises(cli.UploadError): cli.prepare_target(FakeAPI(),{'project_id':'p','chapter':{'label':'1','name':'Wanted','order':1}},{},lambda:None)

    def test_draft_mismatch_never_silently_overwrites_other_pages(self):
        expected=[{'key':'wanted','position':1}]
        with self.assertRaises(cli.UploadError): cli.validate_draft({'pages':[{'upload_key':'foreign','position':1}]},expected)
        with self.assertRaises(cli.UploadError): cli.validate_draft({'pages':[{'upload_key':'wanted','position':2}]},expected)
        cli.validate_draft({'pages':[{'upload_key':'wanted','position':1}]},expected)

if __name__ == '__main__':
    import unittest.mock
    unittest.main()
