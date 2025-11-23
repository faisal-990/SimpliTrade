CREATE TABLE "User" (
  "ID" int PRIMARY KEY,
  "Name" varchar(100) NOT NULL,
  "Email" varchar(100) UNIQUE NOT NULL,
  "Password" varchar NOT NULL,
  "LastLoginAt" timestamp,
  "Balance" numeric(15,2) DEFAULT 100000,
  "CreatedAt" timestamp,
  "UpdatedAt" timestamp,
  "DeletedAt" timestamp
);

CREATE TABLE "Investor" (
  "ID" int PRIMARY KEY,
  "Bio" text,
  "Strategy" text,
  "CreatedAt" timestamp,
  "UpdatedAt" timestamp
);

CREATE TABLE "Performance" (
  "ID" int PRIMARY KEY,
  "InvestorID" int UNIQUE NOT NULL,
  "ROI" numeric(5,2) NOT NULL,
  "Rank" int NOT NULL,
  "LastUpdate" timestamp NOT NULL,
  "CreatedAt" timestamp,
  "UpdatedAt" timestamp
);

CREATE TABLE "Follow" (
  "ID" int PRIMARY KEY,
  "InvestorID" int NOT NULL,
  "FollowerID" int NOT NULL,
  "CreatedAt" timestamp
);

CREATE TABLE "Stock" (
  "ID" int PRIMARY KEY,
  "Symbol" varchar(25) UNIQUE NOT NULL,
  "Name" varchar(100) NOT NULL,
  "Exchange" varchar(50),
  "CurrentPrice" numeric(10,2) NOT NULL,
  "Currency" varchar(3) DEFAULT 'USD',
  "CreatedAt" timestamp,
  "UpdatedAt" timestamp,
  "DeletedAt" timestamp
);

CREATE TABLE "StockPrice" (
  "ID" int PRIMARY KEY,
  "StockID" int NOT NULL,
  "Timestamp" timestamp NOT NULL,
  "Open" numeric(10,2),
  "Close" numeric(10,2),
  "High" numeric(10,2),
  "Low" numeric(10,2),
  "Volume" bigint,
  "Interval" varchar(10) NOT NULL,
  "CreatedAt" timestamp
);

CREATE TABLE "Trade" (
  "ID" int PRIMARY KEY,
  "UserID" int NOT NULL,
  "StockID" int NOT NULL,
  "Type" varchar(4) NOT NULL,
  "Quantity" numeric(10,2) NOT NULL,
  "Price" numeric(10,2) NOT NULL,
  "TotalValue" numeric(15,2),
  "ExecutedAt" timestamp NOT NULL,
  "Status" varchar(20) DEFAULT 'executed',
  "InvestorID" int,
  "CreatedAt" timestamp,
  "DeletedAt" timestamp
);

CREATE TABLE "Holding" (
  "ID" int PRIMARY KEY,
  "UserID" int NOT NULL,
  "StockID" int NOT NULL,
  "Quantity" numeric(10,2) NOT NULL,
  "AvgPrice" numeric(10,2) NOT NULL,
  "CreatedAt" timestamp,
  "UpdatedAt" timestamp
);

CREATE UNIQUE INDEX ON "Holding" ("UserID", "StockID");

COMMENT ON COLUMN "Trade"."Type" IS '"buy" or "sell"';

COMMENT ON COLUMN "Trade"."TotalValue" IS 'Price * Quantity';

COMMENT ON COLUMN "Trade"."InvestorID" IS 'Nullable, for copy trades';

ALTER TABLE "Follow" ADD FOREIGN KEY ("InvestorID") REFERENCES "Investor" ("ID");

ALTER TABLE "Follow" ADD FOREIGN KEY ("FollowerID") REFERENCES "User" ("ID");

ALTER TABLE "StockPrice" ADD FOREIGN KEY ("StockID") REFERENCES "Stock" ("ID");

ALTER TABLE "Trade" ADD FOREIGN KEY ("UserID") REFERENCES "User" ("ID");

ALTER TABLE "Trade" ADD FOREIGN KEY ("StockID") REFERENCES "Stock" ("ID");

ALTER TABLE "Trade" ADD FOREIGN KEY ("InvestorID") REFERENCES "Investor" ("ID");

ALTER TABLE "User" ADD FOREIGN KEY ("ID") REFERENCES "Investor" ("ID");

ALTER TABLE "Investor" ADD FOREIGN KEY ("ID") REFERENCES "Performance" ("InvestorID");

ALTER TABLE "Holding" ADD FOREIGN KEY ("UserID") REFERENCES "User" ("ID");

ALTER TABLE "Holding" ADD FOREIGN KEY ("StockID") REFERENCES "Stock" ("ID");
