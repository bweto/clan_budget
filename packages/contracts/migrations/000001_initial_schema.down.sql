-- Drop policies
DROP POLICY IF EXISTS rules_manage_policy ON recurring_rules;
DROP POLICY IF EXISTS rules_select_policy ON recurring_rules;

DROP POLICY IF EXISTS transactions_update_delete_policy ON transactions;
DROP POLICY IF EXISTS transactions_insert_policy ON transactions;
DROP POLICY IF EXISTS transactions_select_policy ON transactions;

DROP POLICY IF EXISTS group_members_manage_policy ON group_members;
DROP POLICY IF EXISTS group_members_select_policy ON group_members;

DROP POLICY IF EXISTS group_update_delete_policy ON family_groups;
DROP POLICY IF EXISTS group_insert_policy ON family_groups;
DROP POLICY IF EXISTS group_select_policy ON family_groups;

DROP POLICY IF EXISTS users_isolation_policy ON users;

-- Drop tables
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS recurring_rules;
DROP TABLE IF EXISTS group_members;
DROP TABLE IF EXISTS family_groups;
DROP TABLE IF EXISTS users;

-- Drop extension
DROP EXTENSION IF EXISTS "uuid-ossp";
