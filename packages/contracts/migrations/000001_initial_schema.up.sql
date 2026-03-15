-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- USERS TABLE
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- FAMILY GROUPS TABLE
CREATE TABLE family_groups (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- GROUP MEMBERS TABLE
-- Roles can be 'admin' or 'member'
CREATE TABLE group_members (
    group_id UUID REFERENCES family_groups(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL DEFAULT 'member',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (group_id, user_id)
);

-- RECURRING RULES TABLE (Fixed expenses/incomes)
CREATE TABLE recurring_rules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    group_id UUID REFERENCES family_groups(id) ON DELETE CASCADE,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    type VARCHAR(50) NOT NULL CHECK (type IN ('income', 'expense')),
    amount DECIMAL(12, 2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    description TEXT,
    frequency VARCHAR(50) NOT NULL CHECK (frequency IN ('daily', 'weekly', 'monthly', 'yearly')),
    start_date DATE NOT NULL,
    end_date DATE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- TRANSACTIONS TABLE
CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    group_id UUID REFERENCES family_groups(id) ON DELETE CASCADE, -- Optional: Can be null for personal transactions
    recurring_rule_id UUID REFERENCES recurring_rules(id) ON DELETE SET NULL,
    type VARCHAR(50) NOT NULL CHECK (type IN ('income', 'expense')),
    amount DECIMAL(12, 2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    description TEXT,
    date DATE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- INDEXES
CREATE INDEX idx_transactions_user_id ON transactions(user_id);
CREATE INDEX idx_transactions_group_id ON transactions(group_id);
CREATE INDEX idx_transactions_date ON transactions(date);
CREATE INDEX idx_recurring_rules_group_id ON recurring_rules(group_id);

-- ==========================================
-- ROW-LEVEL SECURITY (RLS) POLICIES
-- ==========================================

-- Enable RLS on all tables
ALTER TABLE users ENABLE ROW LEVEL SECURITY;
ALTER TABLE family_groups ENABLE ROW LEVEL SECURITY;
ALTER TABLE group_members ENABLE ROW LEVEL SECURITY;
ALTER TABLE recurring_rules ENABLE ROW LEVEL SECURITY;
ALTER TABLE transactions ENABLE ROW LEVEL SECURITY;

-- Note on context:
-- In the Go application, before querying, we should set the Postgres local variable for the current user:
-- SET LOCAL app.current_user_id = 'uuid-of-user';
-- The following policies depend on this variable or explicit filtering.
-- To allow the API to work purely with explicit `WHERE user_id = $1` filters while still enforcing security,
-- the policies below use standard role-based logic assuming auth through an API layer.

-- Users: A user can only read and update their own record
CREATE POLICY users_isolation_policy ON users
    USING (id = current_setting('app.current_user_id', true)::UUID);

-- Family Groups: A user can see groups they created OR groups they are members of
CREATE POLICY group_select_policy ON family_groups FOR SELECT
    USING (
        id IN (
            SELECT group_id FROM group_members WHERE user_id = current_setting('app.current_user_id', true)::UUID
        )
    );

CREATE POLICY group_insert_policy ON family_groups FOR INSERT
    WITH CHECK (created_by = current_setting('app.current_user_id', true)::UUID);

CREATE POLICY group_update_delete_policy ON family_groups FOR ALL
    USING (
        id IN (
            SELECT group_id FROM group_members
            WHERE user_id = current_setting('app.current_user_id', true)::UUID AND role = 'admin'
        )
    );

-- Group Members: Users can see members of their own groups
CREATE POLICY group_members_select_policy ON group_members FOR SELECT
    USING (
        group_id IN (
            SELECT group_id FROM group_members WHERE user_id = current_setting('app.current_user_id', true)::UUID
        )
    );

-- Group Members Management: Only admins can manage members
CREATE POLICY group_members_manage_policy ON group_members FOR ALL
    USING (
        group_id IN (
            SELECT group_id FROM group_members 
            WHERE user_id = current_setting('app.current_user_id', true)::UUID AND role = 'admin'
        )
    );

-- Transactions: A user can see their own personal transactions AND all transactions in their groups
CREATE POLICY transactions_select_policy ON transactions FOR SELECT
    USING (
        user_id = current_setting('app.current_user_id', true)::UUID 
        OR 
        group_id IN (
            SELECT group_id FROM group_members WHERE user_id = current_setting('app.current_user_id', true)::UUID
        )
    );

CREATE POLICY transactions_insert_policy ON transactions FOR INSERT
    WITH CHECK (
        user_id = current_setting('app.current_user_id', true)::UUID
        AND (
            group_id IS NULL OR 
            group_id IN (
                SELECT group_id FROM group_members WHERE user_id = current_setting('app.current_user_id', true)::UUID
            )
        )
    );

CREATE POLICY transactions_update_delete_policy ON transactions FOR UPDATE
    USING (user_id = current_setting('app.current_user_id', true)::UUID); -- Only owner can edit/delete

-- Recurring Rules: Users can see rules for their groups
CREATE POLICY rules_select_policy ON recurring_rules FOR SELECT
    USING (
        group_id IN (
            SELECT group_id FROM group_members WHERE user_id = current_setting('app.current_user_id', true)::UUID
        )
    );

CREATE POLICY rules_manage_policy ON recurring_rules FOR ALL
    USING (
        group_id IN (
            SELECT group_id FROM group_members WHERE user_id = current_setting('app.current_user_id', true)::UUID
        )
    );
