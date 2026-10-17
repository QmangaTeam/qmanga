CREATE UNIQUE INDEX team_invitation_email_idx ON team_invitations(team_id,email);
CREATE INDEX team_membership_user_idx ON team_memberships(user_id);
ALTER TABLE team_memberships ADD CONSTRAINT team_role_check CHECK(role IN ('member','editor','owner'));
ALTER TABLE team_invitations ADD CONSTRAINT invitation_role_check CHECK(role IN ('member','editor'));
