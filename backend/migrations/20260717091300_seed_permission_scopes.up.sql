INSERT INTO
    permission_scopes (
        code,
        resource,
        action,
        scope_level,
        description,
        is_assignable
    )
VALUES (
        'organization.read',
        'organization',
        'read',
        'ORGANIZATION',
        'Read organization details.',
        TRUE
    ),
    (
        'organization.update',
        'organization',
        'update',
        'ORGANIZATION',
        'Update organization settings.',
        TRUE
    ),
    (
        'organization.archive',
        'organization',
        'archive',
        'ORGANIZATION',
        'Archive an organization.',
        TRUE
    ),
    (
        'organization.owners.manage',
        'organization',
        'owners.manage',
        'ORGANIZATION',
        'Assign and remove organization owners.',
        FALSE
    ),
    (
        'stores.read',
        'stores',
        'read',
        'ORGANIZATION',
        'Read organization stores.',
        TRUE
    ),
    (
        'stores.create',
        'stores',
        'create',
        'ORGANIZATION',
        'Create organization stores.',
        TRUE
    ),
    (
        'stores.update',
        'stores',
        'update',
        'ORGANIZATION',
        'Update organization stores.',
        TRUE
    ),
    (
        'stores.status.update',
        'stores',
        'status.update',
        'ORGANIZATION',
        'Change store lifecycle status.',
        TRUE
    ),
    (
        'members.read',
        'members',
        'read',
        'ORGANIZATION',
        'Read organization memberships.',
        TRUE
    ),
    (
        'members.invite',
        'members',
        'invite',
        'ORGANIZATION',
        'Invite organization members.',
        TRUE
    ),
    (
        'members.status.update',
        'members',
        'status.update',
        'ORGANIZATION',
        'Change membership lifecycle status.',
        TRUE
    ),
    (
        'members.remove',
        'members',
        'remove',
        'ORGANIZATION',
        'Remove organization memberships.',
        TRUE
    ),
    (
        'invitations.read',
        'invitations',
        'read',
        'ORGANIZATION',
        'Read organization invitations.',
        TRUE
    ),
    (
        'invitations.manage',
        'invitations',
        'manage',
        'ORGANIZATION',
        'Create and revoke organization invitations.',
        TRUE
    ),
    (
        'scopes.read',
        'scopes',
        'read',
        'ORGANIZATION',
        'Read the platform scope catalog.',
        TRUE
    ),
    (
        'roles.read',
        'roles',
        'read',
        'ORGANIZATION',
        'Read organization roles.',
        TRUE
    ),
    (
        'roles.create',
        'roles',
        'create',
        'ORGANIZATION',
        'Create custom roles.',
        TRUE
    ),
    (
        'roles.update',
        'roles',
        'update',
        'ORGANIZATION',
        'Update mutable roles.',
        TRUE
    ),
    (
        'roles.status.update',
        'roles',
        'status.update',
        'ORGANIZATION',
        'Activate or deactivate mutable roles.',
        TRUE
    ),
    (
        'roles.assign',
        'roles',
        'assign',
        'ORGANIZATION',
        'Assign roles to memberships.',
        TRUE
    ),
    (
        'audit.read',
        'audit',
        'read',
        'ORGANIZATION',
        'Read organization security audit events.',
        TRUE
    ),
    (
        'products.read',
        'products',
        'read',
        'ORGANIZATION',
        'Read organization products.',
        TRUE
    ),
    (
        'products.create',
        'products',
        'create',
        'ORGANIZATION',
        'Create organization products.',
        TRUE
    ),
    (
        'products.update',
        'products',
        'update',
        'ORGANIZATION',
        'Update organization products.',
        TRUE
    ),
    (
        'products.status.update',
        'products',
        'status.update',
        'ORGANIZATION',
        'Change product lifecycle status.',
        TRUE
    ),
    (
        'categories.read',
        'categories',
        'read',
        'ORGANIZATION',
        'Read organization categories.',
        TRUE
    ),
    (
        'categories.create',
        'categories',
        'create',
        'ORGANIZATION',
        'Create organization categories.',
        TRUE
    ),
    (
        'categories.update',
        'categories',
        'update',
        'ORGANIZATION',
        'Update organization categories.',
        TRUE
    ),
    (
        'categories.status.update',
        'categories',
        'status.update',
        'ORGANIZATION',
        'Change category lifecycle status.',
        TRUE
    ),
    (
        'payment_methods.manage',
        'payment_methods',
        'manage',
        'ORGANIZATION',
        'Configure organization payment methods.',
        TRUE
    ),
    (
        'catalog.read',
        'catalog',
        'read',
        'STORE',
        'Read the active store catalog.',
        TRUE
    ),
    (
        'inventory.read',
        'inventory',
        'read',
        'STORE',
        'Read store inventory.',
        TRUE
    ),
    (
        'inventory.entries.create',
        'inventory',
        'entries.create',
        'STORE',
        'Record store inventory entries.',
        TRUE
    ),
    (
        'inventory.adjustments.create',
        'inventory',
        'adjustments.create',
        'STORE',
        'Record manual store inventory adjustments.',
        TRUE
    ),
    (
        'inventory.movements.read',
        'inventory',
        'movements.read',
        'STORE',
        'Read store inventory movements.',
        TRUE
    ),
    (
        'sales.read',
        'sales',
        'read',
        'STORE',
        'Read store sales.',
        TRUE
    ),
    (
        'sales.create',
        'sales',
        'create',
        'STORE',
        'Create store sales.',
        TRUE
    ),
    (
        'sales.items.manage',
        'sales',
        'items.manage',
        'STORE',
        'Manage items in open store sales.',
        TRUE
    ),
    (
        'sales.cancel',
        'sales',
        'cancel',
        'STORE',
        'Cancel store sales.',
        TRUE
    ),
    (
        'sales.checkout',
        'sales',
        'checkout',
        'STORE',
        'Complete store sale checkout.',
        TRUE
    ),
    (
        'payment_methods.read',
        'payment_methods',
        'read',
        'STORE',
        'Read enabled store payment methods.',
        TRUE
    ),
    (
        'payments.read',
        'payments',
        'read',
        'STORE',
        'Read store payment transactions.',
        TRUE
    ),
    (
        'fiscal.read',
        'fiscal',
        'read',
        'STORE',
        'Read store fiscal documents.',
        TRUE
    ),
    (
        'receipts.read',
        'receipts',
        'read',
        'STORE',
        'Read store receipts.',
        TRUE
    );