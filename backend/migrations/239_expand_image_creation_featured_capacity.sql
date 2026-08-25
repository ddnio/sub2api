ALTER TABLE image_creation_templates
    DROP CONSTRAINT IF EXISTS image_creation_templates_home_position_check;

ALTER TABLE image_creation_templates
    ADD CONSTRAINT image_creation_templates_home_position_check
    CHECK (home_position BETWEEN 1 AND 20);
