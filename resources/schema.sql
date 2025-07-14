-- Table : user
CREATE TABLE IF NOT EXISTS `user` (
    user_id      INT AUTO_INCREMENT PRIMARY KEY,
    lastname     VARCHAR(100) NOT NULL,
    firstname    VARCHAR(100) NOT NULL,
    email        VARCHAR(255) NOT NULL UNIQUE,
    created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   DATETIME DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP,
    deleted_at   DATETIME DEFAULT NULL
) ENGINE=InnoDB; 

-- Table : calendar
CREATE TABLE IF NOT EXISTS `calendar` (
    calendar_id  INT AUTO_INCREMENT PRIMARY KEY,
    title        VARCHAR(200) NOT NULL,
    description  TEXT,
    created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   DATETIME DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP,
    deleted_at   DATETIME DEFAULT NULL
) ENGINE=InnoDB;

-- Table : user_calendar
CREATE TABLE IF NOT EXISTS `user_calendar` (
    user_calendar_id INT AUTO_INCREMENT PRIMARY KEY,
    user_id          INT NOT NULL,
    calendar_id      INT NOT NULL,
    created_at       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at       DATETIME DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP,
    deleted_at       DATETIME DEFAULT NULL,
    CONSTRAINT uc_user_calendar UNIQUE (user_id, calendar_id),
    CONSTRAINT fk_user_calendar_user FOREIGN KEY (user_id) REFERENCES `user`(user_id)
        ON DELETE CASCADE,
    CONSTRAINT fk_user_calendar_calendar FOREIGN KEY (calendar_id) REFERENCES `calendar`(calendar_id)
        ON DELETE CASCADE
) ENGINE=InnoDB;

-- Table : event
CREATE TABLE IF NOT EXISTS `event` (
    event_id     INT AUTO_INCREMENT PRIMARY KEY,
    title        VARCHAR(200) NOT NULL,
    description  TEXT,
    start        DATETIME NOT NULL,
    duration     INT NOT NULL,
    canceled     BOOL NOT NULL DEFAULT FALSE,
    created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   DATETIME DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP,
    deleted_at   DATETIME DEFAULT NULL
) ENGINE=InnoDB;

-- Table : calendar_event
CREATE TABLE IF NOT EXISTS `calendar_event` (
    calendar_event_id INT AUTO_INCREMENT PRIMARY KEY,
    calendar_id       INT NOT NULL,
    event_id          INT NOT NULL,
    created_at        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at        DATETIME DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP,
    deleted_at        DATETIME DEFAULT NULL,
    CONSTRAINT fk_calendar_event_calendar FOREIGN KEY (calendar_id) REFERENCES `calendar`(calendar_id)
        ON DELETE CASCADE,
    CONSTRAINT fk_calendar_event_event FOREIGN KEY (event_id) REFERENCES `event`(event_id)
        ON DELETE CASCADE
) ENGINE=InnoDB;

-- Table : user_password
CREATE TABLE IF NOT EXISTS `user_password` (
    user_password_id INT AUTO_INCREMENT PRIMARY KEY,
    user_id          INT NOT NULL,
    password_hash    VARCHAR(500) NOT NULL,
    expired_at       DATETIME DEFAULT NULL,
    created_at       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at       DATETIME DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP,
    deleted_at       DATETIME DEFAULT NULL,
    CONSTRAINT fk_user_password_user FOREIGN KEY (user_id) REFERENCES `user`(user_id)
        ON DELETE CASCADE
) ENGINE=InnoDB;

-- Table : roles
CREATE TABLE IF NOT EXISTS `roles` (
    role_id      INT AUTO_INCREMENT PRIMARY KEY,
    name         VARCHAR(50) NOT NULL UNIQUE,
    description  TEXT,
    created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   DATETIME DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP,
    deleted_at   DATETIME DEFAULT NULL
) ENGINE=InnoDB;

-- Table : user_roles
CREATE TABLE IF NOT EXISTS `user_roles` (
    user_roles_id INT AUTO_INCREMENT PRIMARY KEY,
    user_id       INT NOT NULL,
    role_id       INT NOT NULL,
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    DATETIME DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP,
    deleted_at    DATETIME DEFAULT NULL,
    CONSTRAINT uc_user_roles UNIQUE (user_id, role_id),
    CONSTRAINT fk_user_roles_user FOREIGN KEY (user_id) REFERENCES `user`(user_id)
        ON DELETE CASCADE,
    CONSTRAINT fk_user_roles_role FOREIGN KEY (role_id) REFERENCES `roles`(role_id)
        ON DELETE CASCADE
) ENGINE=InnoDB;

-- Table : user_session
CREATE TABLE IF NOT EXISTS `user_session` (
    user_session_id INT AUTO_INCREMENT PRIMARY KEY,
    user_id         INT NOT NULL,
    session_token   VARCHAR(500) NOT NULL UNIQUE,
    refresh_token   VARCHAR(500) DEFAULT NULL,
    expires_at      DATETIME NOT NULL,
    device_info     VARCHAR(255) DEFAULT NULL,
    ip_address      VARCHAR(45) DEFAULT NULL,
    location        VARCHAR(255) DEFAULT NULL,
    is_active       BOOL NOT NULL DEFAULT TRUE,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP,
    deleted_at      DATETIME DEFAULT NULL,
    CONSTRAINT fk_user_session_user FOREIGN KEY (user_id) REFERENCES `user`(user_id)
        ON DELETE CASCADE
) ENGINE=InnoDB;

-- ===== RÔLES PRÉVUS PAR L'APPLICATION =====

-- Rôle administrateur (accès complet à toutes les fonctionnalités)
INSERT INTO `roles` (name, description, created_at) VALUES 
('admin', 'Administrateur avec tous les droits sur l''application', NOW())
ON DUPLICATE KEY UPDATE updated_at = NOW();

-- Rôle utilisateur standard (accès limité à ses propres données)
INSERT INTO `roles` (name, description, created_at) VALUES 
('user', 'Utilisateur standard avec accès à ses propres calendriers et événements', NOW())
ON DUPLICATE KEY UPDATE updated_at = NOW();

-- Rôle modérateur (permissions intermédiaires)
INSERT INTO `roles` (name, description, created_at) VALUES 
('moderator', 'Modérateur avec permissions étendues de gestion', NOW())
ON DUPLICATE KEY UPDATE updated_at = NOW();

-- Rôle éditeur (permissions de création et modification)
INSERT INTO `roles` (name, description, created_at) VALUES 
('editor', 'Éditeur avec permissions de création et modification de contenu', NOW())
ON DUPLICATE KEY UPDATE updated_at = NOW();

-- ===== ADMINISTRATEUR PAR DÉFAUT =====

-- Création de l'utilisateur administrateur par défaut
INSERT INTO `user` (lastname, firstname, email, created_at) VALUES 
('Administrateur', 'GoLendar', 'admin@golendar.com', NOW())
ON DUPLICATE KEY UPDATE updated_at = NOW();

-- Récupération de l'ID de l'utilisateur admin et du rôle admin
SET @admin_user_id = (SELECT user_id FROM `user` WHERE email = 'admin@golendar.com' LIMIT 1);
SET @admin_role_id = (SELECT role_id FROM `roles` WHERE name = 'admin' LIMIT 1);

-- Création du mot de passe hashé pour l'administrateur (mot de passe: Admin123!)
-- Le hash bcrypt correspond au mot de passe "Admin123!" avec un coût de 10
INSERT INTO `user_password` (user_id, password_hash, created_at) VALUES 
(@admin_user_id, '$2a$10$TmvM0A32sArLU6XNYnR4BuBMv/HrMtFAfVhBtm//aLSwnrHAb.Q8W', NOW())
ON DUPLICATE KEY UPDATE updated_at = NOW();

-- Attribution du rôle admin à l'utilisateur administrateur
INSERT INTO `user_roles` (user_id, role_id, created_at) VALUES 
(@admin_user_id, @admin_role_id, NOW())
ON DUPLICATE KEY UPDATE updated_at = NOW();