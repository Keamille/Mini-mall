-- 容器首次初始化时执行：按服务创建 database，并授权给应用用户
CREATE DATABASE IF NOT EXISTS minimall_user CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE DATABASE IF NOT EXISTS minimall_product CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE DATABASE IF NOT EXISTS minimall_inventory CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE DATABASE IF NOT EXISTS minimall_cart CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE DATABASE IF NOT EXISTS minimall_order CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE DATABASE IF NOT EXISTS minimall_payment CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE DATABASE IF NOT EXISTS minimall_coupon CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE DATABASE IF NOT EXISTS minimall_notification CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE DATABASE IF NOT EXISTS minimall_search CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

GRANT ALL PRIVILEGES ON minimall_user.* TO 'minimall'@'%';
GRANT ALL PRIVILEGES ON minimall_product.* TO 'minimall'@'%';
GRANT ALL PRIVILEGES ON minimall_inventory.* TO 'minimall'@'%';
GRANT ALL PRIVILEGES ON minimall_cart.* TO 'minimall'@'%';
GRANT ALL PRIVILEGES ON minimall_order.* TO 'minimall'@'%';
GRANT ALL PRIVILEGES ON minimall_payment.* TO 'minimall'@'%';
GRANT ALL PRIVILEGES ON minimall_coupon.* TO 'minimall'@'%';
GRANT ALL PRIVILEGES ON minimall_notification.* TO 'minimall'@'%';
GRANT ALL PRIVILEGES ON minimall_search.* TO 'minimall'@'%';
FLUSH PRIVILEGES;
