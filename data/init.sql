DROP DATABASE IF EXISTS invoice_db;
CREATE DATABASE invoice_db;
USE invoice_db;

DROP TABLE IF EXISTS invoice;

CREATE TABLE IF NOT EXISTS invoice (
  invoice_id    INT NOT NULL PRIMARY KEY AUTO_INCREMENT,
  company_id    INT NOT NULL,
  issue_date    DATE NOT NULL,
  amount        INT NOT NULL,
  fee           INT NOT NULL,
  fee_rate      DECIMAL(3, 2) NOT NULL,
  tax           INT NOT NULL,
  tax_rate      DECIMAL(3, 2) NOT NULL,
  total         INT NOT NULL,
  due_date      DATE NOT NULL,
  status        ENUM("unprocessed", "processing", "paid", "error") NOT NULL,
  CONSTRAINT `total_check` CHECK ((`amount` + `fee` + `tax` = `total`)),
  CONSTRAINT `fee_check` CHECK ((`amount` * `fee_rate` = `fee`)),
  CONSTRAINT `tax_check` CHECK ((`fee` * `tax_rate` = `tax`))
);

INSERT INTO invoice (company_id, issue_date, amount, fee, fee_rate, tax, tax_rate, total, due_date, status) VALUES (1, "2024-11-01", 10000, 400, 0.04, 40, 0.10, 10440, "2024-12-01", "unprocessed");
INSERT INTO invoice (company_id, issue_date, amount, fee, fee_rate, tax, tax_rate, total, due_date, status) VALUES (1, "2024-10-01", 5000, 200, 0.04, 20, 0.10, 5220, "2024-11-01", "processing");
INSERT INTO invoice (company_id, issue_date, amount, fee, fee_rate, tax, tax_rate, total, due_date, status) VALUES (1, "2024-07-01", 20000, 800, 0.04, 80, 0.10, 20880, "2024-08-01", "paid");
INSERT INTO invoice (company_id, issue_date, amount, fee, fee_rate, tax, tax_rate, total, due_date, status) VALUES (2, "2024-04-01", 5000, 200, 0.04, 20, 0.10, 5220, "2024-11-01", "error");
