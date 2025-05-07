-- Drop all tables related to the sales process in reverse order to avoid foreign key conflicts

-- First drop the link tables
DROP TABLE IF EXISTS process_invoices;
DROP TABLE IF EXISTS process_deliveries;
DROP TABLE IF EXISTS process_purchase_orders;
DROP TABLE IF EXISTS process_sales_orders;
DROP TABLE IF EXISTS process_quotations;

-- Drop the sales process table
DROP TABLE IF EXISTS sales_processes;

-- Drop the payment tables
DROP TABLE IF EXISTS payments;

-- Drop the invoice tables
DROP TABLE IF EXISTS invoice_items;
DROP TABLE IF EXISTS invoices;

-- Drop the delivery tables
DROP TABLE IF EXISTS delivery_items;
DROP TABLE IF EXISTS deliveries;

-- Drop the purchase order tables
DROP TABLE IF EXISTS purchase_order_items;
DROP TABLE IF EXISTS purchase_orders;

-- Drop the sales order tables
DROP TABLE IF EXISTS sales_order_items;
DROP TABLE IF EXISTS sales_orders;

-- Drop the quotation tables
DROP TABLE IF EXISTS quotation_items;
DROP TABLE IF EXISTS quotations;