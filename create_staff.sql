INSERT INTO staff_members (id, name, email, phone, password, role, is_active, created_at)
VALUES (
    gen_random_uuid(),
    'surajit', 
    's@example.com', 
    '898989',
    '$2a$10$VGSh5VUNsbppCiRpA.boUOpHVXJLkmSJRqYGKTCSoOrWJuH6VyINS', 
    'MANAGER', 
    true,
    NOW()
);