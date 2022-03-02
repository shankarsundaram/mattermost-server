DO $$
DECLARE 
    column_exist boolean := false;
BEGIN
SELECT count(*) != 0 INTO column_exist
    FROM information_schema.columns
    WHERE table_name = 'channelmembers'
    AND column_name = 'roles'
    AND NOT data_type = 'varchar(64)';
IF column_exist THEN
    ALTER TABLE channelmembers ALTER COLUMN roles TYPE varchar(64);
END IF;
END $$;
