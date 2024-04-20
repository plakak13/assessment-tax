DROP TYPE IF EXISTS tax_allowance_type; 
CREATE TYPE tax_allowance_type AS ENUM ('donation','k-receipt','personal'); 

CREATE TABLE IF NOT EXISTS tax_rate (
id SERIAL PRIMARY KEY,
lower_bound_income DECIMAL (10,2) NOT NULL,
tax_rate DECIMAL (10,2) NOT NULL,
created_at TIMESTAMP NOT NULL DEFAULT now(),
updated_at TIMESTAMP NULL DEFAULT NULL); 

CREATE TABLE IF NOT EXISTS tax_deduction (
id SERIAL PRIMARY KEY,
max_deduction_amount DECIMAL (18,2),
default_amount DECIMAL (18,2),
admin_override_max DECIMAL (18,2),
min_amount DECIMAL (18,2),
tax_allowance_type tax_allowance_type NOT NULL,
created_at TIMESTAMP NOT NULL DEFAULT now(),
updated_at TIMESTAMP NULL DEFAULT NULL); 

CREATE OR REPLACE FUNCTION update_updated_at_column () 
RETURNS TRIGGER AS $$ 
BEGIN 
	NEW."updated_at"=CURRENT_TIMESTAMP; 
	RETURN NEW; 
END;
$$ LANGUAGE plpgsql; 

CREATE TRIGGER update_taxrate_updated_at BEFORE 
UPDATE ON tax_rate FOR EACH ROW EXECUTE FUNCTION update_updated_at_column (); 

CREATE TRIGGER update_taxdeduction_updated_at BEFORE 
UPDATE ON tax_deduction FOR EACH ROW EXECUTE FUNCTION update_updated_at_column (); 

INSERT INTO "tax_rate" ("lower_bound_income","tax_rate","created_at") VALUES 
('0.00','0.00',now()),
('150001.00','10.00',now()),
('500001.00','15.00',now()),
('1000001.00','20.00',now()),
('2000001.00','35.00',now()); 

INSERT INTO "tax_deduction" ("max_deduction_amount","default_amount","admin_override_max","min_amount","tax_allowance_type","created_at","updated_at") VALUES 
('100000.00',NULL,NULL,NULL,'donation',now(),NULL),
('50000.00','50000.00','100000.00','0.00','k-receipt',now(),NULL),
('60000.00','60000.00','100000.00','10000.00','personal',now(),NULL);