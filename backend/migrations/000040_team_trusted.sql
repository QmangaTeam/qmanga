-- Trusted teams publish chapters immediately on submit; moderators review afterwards.
ALTER TABLE teams ADD COLUMN trusted BOOLEAN NOT NULL DEFAULT false;
