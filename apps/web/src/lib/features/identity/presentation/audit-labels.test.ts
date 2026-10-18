import {test,expect} from 'bun:test';
import {auditActorLabel,auditCoverValue,coverAuditCopy} from './audit-labels';

test('cover actions and removal state have readable Kazakh and Russian labels',()=>{
 expect(coverAuditCopy('kk').select_cover).toBe('Мұқабаны таңдау');
 expect(coverAuditCopy('kk').remove_cover).toBe('Мұқабаны алып тастау');
 expect(coverAuditCopy('ru').select_cover).toBe('Выбор обложки');
 expect(coverAuditCopy('ru').remove_cover).toBe('Удаление обложки');
 expect(auditCoverValue('cover-12345678','ru')).toBe('Обложка · 12345678');
 expect(auditCoverValue('','ru')).toBe('Нет обложки');
 expect(auditCoverValue('','kk')).toBe('Мұқаба жоқ');
 expect(auditCoverValue(undefined,'kk')).toBe('—');
});

test('unnamed actors remain identifiable without account emails',()=>{
 expect(auditActorLabel('','account-12345678','kk')).toBe('Аккаунт · 12345678');
 expect(auditActorLabel('   ','account-12345678','ru')).toBe('Аккаунт · 12345678');
 expect(auditActorLabel('Moderator','account-12345678','ru')).toBe('Moderator');
 expect(auditActorLabel('','','ru')).toBe('Аккаунт недоступен');
});
