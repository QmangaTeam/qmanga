import { AppFailure } from '$lib/shared/domain/failure';
/** Local-only PDF rasterization. Only resulting images enter the normal upload pipeline. */
const MAX_INPUT = 50 * 1024 * 1024;
const MAX_OUTPUT = 250 * 1024 * 1024;
const MAX_PAGE_PIXELS = 12_000_000;
const MAX_TOTAL_PIXELS = 120_000_000;

export function validatePdfInput(bytes: number, prefix: string): void {
  if (bytes <= 0 || bytes > MAX_INPUT) throw new AppFailure('import_size');
  if (!prefix.startsWith('%PDF-')) throw new AppFailure('import_invalid');
}
export function validatePdfPageCount(count: number): void {
  if (!Number.isInteger(count) || count < 1 || count > 100) throw new AppFailure('import_size');
}
export function pdfPageSize(width: number, height: number) {
  if (!Number.isFinite(width) || !Number.isFinite(height) || width <= 0 || height <= 0) throw new AppFailure('import_invalid');
  const scale = 1600 / width;
  const outputHeight = Math.ceil(height * scale);
  if (outputHeight < 1 || outputHeight > 7500 || 1600 * outputHeight > MAX_PAGE_PIXELS) throw new AppFailure('import_size');
  return { width: 1600, height: outputHeight, scale };
}

export async function importPdf(file: File, options: { signal?: AbortSignal; onProgress?: (done: number, total: number) => void } = {}): Promise<File[]> {
  validatePdfInput(file.size, await file.slice(0, 8).text());
  if (options.signal?.aborted) throw new AppFailure('upload_cancelled');
  // Import lazily: regular image uploads and SSR do not load the PDF decoder.
  const [pdfjs, { default: workerUrl }] = await Promise.all([
    import('pdfjs-dist'), import('pdfjs-dist/build/pdf.worker.min.mjs?url')
  ]);
  if (options.signal?.aborted) throw new AppFailure('upload_cancelled');
  const worker = new Worker(workerUrl, { type: 'module' });
  const pdfWorker = pdfjs.PDFWorker.create({ port: worker });
  let task: ReturnType<typeof pdfjs.getDocument> | undefined;
  let render: { cancel(): void } | undefined;
  let stopped: Error | undefined;
  let rejectStop!: (error: Error) => void;
  const stopPromise = new Promise<never>((_, reject) => { rejectStop = reject; });
  const stop = (error: Error) => {
    if (stopped) return;
    stopped = error;
    render?.cancel();
    worker.terminate();
    rejectStop(error);
  };
  const abort = () => stop(new AppFailure('upload_cancelled'));
  options.signal?.addEventListener('abort', abort, { once: true });
  const timeout = setTimeout(() => stop(new AppFailure('pdf_timeout')), 60_000);
  if (options.signal?.aborted) abort();
  const run = async () => {
    const data = new Uint8Array(await file.arrayBuffer());
    if (stopped) throw stopped;
    task = pdfjs.getDocument({ data, worker: pdfWorker, enableXfa: false,
      cMapUrl: '/_pdf-assets/cmaps/', cMapPacked: true, standardFontDataUrl: '/_pdf-assets/standard_fonts/',
      wasmUrl: '/_pdf-assets/wasm/', iccUrl: '/_pdf-assets/iccs/', useSystemFonts: false, disableFontFace: true, useWorkerFetch: false, isImageDecoderSupported: false,
      maxImageSize: 16_000_000, canvasMaxAreaInBytes: 48_000_000, stopAtErrors: true });
    const document = await task.promise;
    validatePdfPageCount(document.numPages);
    // Reject all encrypted documents, even when an empty password happens to work.
    const metadata = await document.getMetadata();
    if ((metadata.info as { EncryptFilterName?: string }).EncryptFilterName) throw new AppFailure('pdf_password');
    options.onProgress?.(0, document.numPages);
    const files: File[] = [];
    let bytes = 0;
    let pixels = 0;
    for (let index = 1; index <= document.numPages; index++) {
      if (stopped) throw stopped;
      const page = await document.getPage(index);
      const base = page.getViewport({ scale: 1 });
      const size = pdfPageSize(base.width, base.height);
      pixels += size.width * size.height;
      if (pixels > MAX_TOTAL_PIXELS) throw new AppFailure('import_size');
      const canvas = window.document.createElement('canvas');
      canvas.width = size.width;
      canvas.height = size.height;
      try {
        render = page.render({ canvas, viewport: page.getViewport({ scale: size.scale }),
          background: 'rgb(255,255,255)', annotationMode: pdfjs.AnnotationMode.DISABLE });
        await (render as ReturnType<typeof page.render>).promise;
        if (stopped) throw stopped;
        const blob = await new Promise<Blob>((resolve, reject) => canvas.toBlob((result) => result ? resolve(result) : reject(new AppFailure('import_invalid')), 'image/png'));
        bytes += blob.size;
        if (blob.size > 25 * 1024 * 1024 || bytes > MAX_OUTPUT) throw new AppFailure('import_size');
        files.push(new File([blob], `pdf-page-${String(index).padStart(3, '0')}.png`, { type: 'image/png', lastModified: 0 }));
        options.onProgress?.(index, document.numPages);
      } finally {
        render = undefined;
        canvas.width = canvas.height = 0;
        page.cleanup();
      }
    }
    return files;
  };
  try {
    return await Promise.race([run(), stopPromise]);
  } catch (error) {
    if (stopped) throw stopped;
    if (error instanceof Error && error.name === 'PasswordException') throw new AppFailure('pdf_password');
    if (error instanceof Error && ['InvalidPDFException', 'UnknownErrorException'].includes(error.name)) throw new AppFailure('import_invalid');
    throw error;
  } finally {
    clearTimeout(timeout);
    options.signal?.removeEventListener('abort', abort);
    render?.cancel();
    worker.terminate();
    // Destroy also releases transferred document buffers and decoder caches.
    if (task) void task.destroy().catch(() => {});
    else pdfWorker.destroy();
  }
}
