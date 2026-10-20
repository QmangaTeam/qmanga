import { describe, expect, test } from 'bun:test';
import { pdfPageSize, validatePdfInput, validatePdfPageCount } from './pdf-import';

describe('bounded PDF import', () => {
 test('rejects empty, huge, and non-PDF input before decoder', () => {
  expect(() => validatePdfInput(0, '%PDF-')).toThrow();
  expect(() => validatePdfInput(51 * 1024 * 1024, '%PDF-')).toThrow();
  expect(() => validatePdfInput(100, '<html>')).toThrow();
  expect(() => validatePdfInput(100, '%PDF-1.7')).not.toThrow();
 });
 test('enforces page count and finite dimensions', () => {
  for (const count of [0,101, Infinity,1.5]) expect(() => validatePdfPageCount(count)).toThrow();
  expect(() => validatePdfPageCount(100)).not.toThrow();
  for (const size of [[0,100], [Infinity,10], [10,NaN], [1,1000000]]) expect(() => pdfPageSize(size[0],size[1])).toThrow();
 });
 test('retains readable 1600px width while bounding long pages', () => {
  expect(pdfPageSize(600,900)).toEqual({width:1600,height:2400,scale:1600/600});
  expect(() => pdfPageSize(600,10000)).toThrow();
 });
});
