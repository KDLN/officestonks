// Storage management service for localStorage validation and versioning
// Ensures data integrity across app updates and handles schema migrations

// Current storage schema version
const STORAGE_VERSION = '1.0.0';
const VERSION_KEY = '_storage_version';
const BACKUP_PREFIX = '_backup_';

// Define expected localStorage schema
const STORAGE_SCHEMA = {
  // Auth data
  'token': { type: 'string', required: false },
  'userId': { type: 'string', required: false },
  'username': { type: 'string', required: false },
  'isAdmin': { type: 'string', required: false },
  
  // UI preferences
  'chatDrawerOpen': { type: 'string', required: false, default: 'false' },
  'theme': { type: 'string', required: false, default: 'dark' },
  'darkMode': { type: 'string', required: false }, // Legacy theme setting
  'newsVisible': { type: 'string', required: false },
  
  // App state
  'lastLoginTime': { type: 'string', required: false },
  'changelogDismissed': { type: 'string', required: false },
  'lastSeenChangelogVersion': { type: 'string', required: false },
  'app_version': { type: 'string', required: false },
  
  // Supabase auth (external)
  'sb-bmsmtdzsexzdnaivieqz-auth-token': { type: 'json', required: false },
  
  // Version tracking
  [VERSION_KEY]: { type: 'string', required: true, default: STORAGE_VERSION }
};

class StorageManager {
  constructor() {
    this.isValidated = false;
    this.migrationLog = [];
  }

  // Initialize storage validation
  async initialize() {
    console.log('📦 Initializing storage manager...');
    
    try {
      await this.validateAndMigrate();
      this.isValidated = true;
      console.log('📦 Storage validation completed successfully');
    } catch (error) {
      console.error('📦 Storage validation failed:', error);
      await this.handleValidationFailure(error);
    }
  }

  // Validate current storage and migrate if needed
  async validateAndMigrate() {
    const currentVersion = localStorage.getItem(VERSION_KEY);
    
    if (!currentVersion) {
      console.log('📦 No storage version found, initializing fresh storage');
      await this.initializeFreshStorage();
      return;
    }

    if (currentVersion !== STORAGE_VERSION) {
      console.log(`📦 Storage version mismatch: ${currentVersion} -> ${STORAGE_VERSION}`);
      await this.migrateStorage(currentVersion, STORAGE_VERSION);
      return;
    }

    // Validate current storage structure
    const validation = this.validateStorageStructure();
    if (!validation.isValid) {
      console.log('📦 Storage structure validation failed:', validation.errors);
      await this.repairStorage(validation);
    }

    console.log('📦 Storage validation passed');
  }

  // Initialize fresh storage with defaults
  async initializeFreshStorage() {
    console.log('📦 Setting up fresh localStorage structure');
    
    // Backup existing data before clearing
    await this.createBackup('pre-init');
    
    // Set up schema with defaults
    Object.entries(STORAGE_SCHEMA).forEach(([key, config]) => {
      if (config.default !== undefined) {
        localStorage.setItem(key, config.default);
      }
    });

    // Set version
    localStorage.setItem(VERSION_KEY, STORAGE_VERSION);
    
    this.migrationLog.push({
      type: 'init',
      timestamp: new Date().toISOString(),
      message: 'Fresh storage initialized'
    });
  }

  // Migrate storage between versions
  async migrateStorage(fromVersion, toVersion) {
    console.log(`📦 Migrating storage from ${fromVersion} to ${toVersion}`);
    
    // Create backup before migration
    await this.createBackup(`pre-migration-${fromVersion}`);
    
    // Version-specific migrations
    if (fromVersion === '0.9.0' && toVersion === '1.0.0') {
      await this.migrateFrom090To100();
    }
    
    // Add more migration paths as needed
    // if (fromVersion === '1.0.0' && toVersion === '1.1.0') {
    //   await this.migrateFrom100To110();
    // }

    // Update version
    localStorage.setItem(VERSION_KEY, toVersion);
    
    this.migrationLog.push({
      type: 'migration',
      from: fromVersion,
      to: toVersion,
      timestamp: new Date().toISOString(),
      message: `Migration completed: ${fromVersion} -> ${toVersion}`
    });
  }

  // Specific migration: 0.9.0 -> 1.0.0
  async migrateFrom090To100() {
    console.log('📦 Running migration: 0.9.0 -> 1.0.0');
    
    // Example: Rename old keys
    const oldThemeKey = 'user-theme-preference';
    const oldTheme = localStorage.getItem(oldThemeKey);
    if (oldTheme) {
      localStorage.setItem('theme', oldTheme);
      localStorage.removeItem(oldThemeKey);
    }

    // Example: Convert data formats
    const oldChatState = localStorage.getItem('chat-drawer-state');
    if (oldChatState === 'visible') {
      localStorage.setItem('chatDrawerOpen', 'true');
    } else if (oldChatState === 'hidden') {
      localStorage.setItem('chatDrawerOpen', 'false');
    }
    localStorage.removeItem('chat-drawer-state');

    // Clean up old keys that are no longer needed
    const deprecatedKeys = [
      'old-user-preferences',
      'legacy-settings',
      'temp-cache'
    ];
    
    deprecatedKeys.forEach(key => {
      if (localStorage.getItem(key)) {
        localStorage.removeItem(key);
        console.log(`📦 Removed deprecated key: ${key}`);
      }
    });
  }

  // Helper function to validate a single schema key
  validateSchemaKey(key, config) {
    const value = localStorage.getItem(key);
    const issues = { errors: [], warnings: [] };

    // Check required fields
    if (config.required && !value) {
      issues.errors.push(`Missing required key: ${key}`);
      return issues;
    }

    // Check data types
    if (value && config.type && !this.validateDataType(value, config.type)) {
      issues.warnings.push(`Invalid type for ${key}: expected ${config.type}`);
    }

    return issues;
  }

  // Helper function to check if a key is expected
  isExpectedKey(key) {
    return STORAGE_SCHEMA[key] || 
           key.startsWith(BACKUP_PREFIX) || 
           (key.startsWith('sb-') && key.includes('-auth-token'));
  }

  // Validate storage structure against schema (simplified)
  validateStorageStructure() {
    const errors = [];
    const warnings = [];

    // Validate schema keys
    Object.entries(STORAGE_SCHEMA).forEach(([key, config]) => {
      const issues = this.validateSchemaKey(key, config);
      errors.push(...issues.errors);
      warnings.push(...issues.warnings);
    });

    // Check for unexpected keys
    Object.keys(localStorage).forEach(key => {
      if (!this.isExpectedKey(key)) {
        warnings.push(`Unexpected localStorage key: ${key}`);
      }
    });

    return {
      isValid: errors.length === 0,
      errors,
      warnings
    };
  }

  // Validate data type
  validateDataType(value, expectedType) {
    switch (expectedType) {
      case 'string':
        return typeof value === 'string';
      case 'number':
        return !isNaN(Number(value));
      case 'boolean':
        return value === 'true' || value === 'false';
      case 'json':
        try {
          JSON.parse(value);
          return true;
        } catch {
          return false;
        }
      default:
        return true;
    }
  }

  // Repair storage issues
  async repairStorage(validation) {
    console.log('📦 Repairing storage issues...');
    
    // Create backup before repair
    await this.createBackup('pre-repair');

    // Fix missing required fields
    validation.errors.forEach(error => {
      if (error.includes('Missing required key:')) {
        const key = error.split(': ')[1];
        const config = STORAGE_SCHEMA[key];
        if (config && config.default !== undefined) {
          localStorage.setItem(key, config.default);
          console.log(`📦 Repaired missing key: ${key} = ${config.default}`);
        }
      }
    });

    // Fix type issues
    validation.warnings.forEach(warning => {
      if (warning.includes('Invalid type for')) {
        const key = warning.split(' ')[3].replace(':', '');
        const config = STORAGE_SCHEMA[key];
        if (config && config.default !== undefined) {
          localStorage.setItem(key, config.default);
          console.log(`📦 Reset invalid value: ${key} = ${config.default}`);
        }
      }
    });

    this.migrationLog.push({
      type: 'repair',
      timestamp: new Date().toISOString(),
      message: `Repaired ${validation.errors.length} errors and ${validation.warnings.length} warnings`
    });
  }

  // Create backup of current localStorage
  async createBackup(suffix) {
    const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
    const backupKey = `${BACKUP_PREFIX}${suffix}_${timestamp}`;
    
    const currentData = {};
    Object.keys(localStorage).forEach(key => {
      if (!key.startsWith(BACKUP_PREFIX)) {
        currentData[key] = localStorage.getItem(key);
      }
    });

    try {
      localStorage.setItem(backupKey, JSON.stringify(currentData));
      console.log(`📦 Created storage backup: ${backupKey}`);
      
      // Clean up old backups (keep only last 3)
      this.cleanupOldBackups();
    } catch (error) {
      console.warn('📦 Failed to create backup:', error);
    }
  }

  // Clean up old backups
  cleanupOldBackups() {
    const backupKeys = Object.keys(localStorage)
      .filter(key => key.startsWith(BACKUP_PREFIX))
      .sort()
      .reverse();

    // Keep only the 3 most recent backups
    if (backupKeys.length > 3) {
      backupKeys.slice(3).forEach(key => {
        localStorage.removeItem(key);
        console.log(`📦 Removed old backup: ${key}`);
      });
    }
  }

  // Handle validation failures
  async handleValidationFailure(error) {
    console.error('📦 Critical storage validation failure:', error);
    
    // Try to restore from backup
    const backups = this.getAvailableBackups();
    if (backups.length > 0) {
      console.log('📦 Attempting to restore from backup...');
      await this.restoreFromBackup(backups[0]);
      return;
    }

    // Last resort: clear corrupted storage and start fresh
    console.log('📦 No backups available, initializing fresh storage');
    localStorage.clear();
    await this.initializeFreshStorage();
  }

  // Get available backups
  getAvailableBackups() {
    return Object.keys(localStorage)
      .filter(key => key.startsWith(BACKUP_PREFIX))
      .sort()
      .reverse();
  }

  // Restore from backup
  async restoreFromBackup(backupKey) {
    try {
      const backupData = JSON.parse(localStorage.getItem(backupKey));
      
      // Clear current storage
      localStorage.clear();
      
      // Restore from backup
      Object.entries(backupData).forEach(([key, value]) => {
        localStorage.setItem(key, value);
      });

      console.log(`📦 Restored from backup: ${backupKey}`);
      
      // Re-validate after restore
      await this.validateAndMigrate();
      
    } catch (error) {
      console.error('📦 Failed to restore from backup:', error);
      throw error;
    }
  }

  // Get storage health status
  getStorageHealth() {
    const validation = this.validateStorageStructure();
    const backups = this.getAvailableBackups();
    
    return {
      isHealthy: validation.errors.length === 0,
      version: localStorage.getItem(VERSION_KEY),
      errors: validation.errors,
      warnings: validation.warnings,
      backupCount: backups.length,
      migrationLog: this.migrationLog,
      lastValidation: new Date().toISOString()
    };
  }

  // Safe localStorage operations
  safeGetItem(key, defaultValue = null) {
    try {
      const value = localStorage.getItem(key);
      return value !== null ? value : defaultValue;
    } catch (error) {
      console.error(`📦 Error reading localStorage key "${key}":`, error);
      return defaultValue;
    }
  }

  safeSetItem(key, value) {
    try {
      localStorage.setItem(key, value);
      return true;
    } catch (error) {
      console.error(`📦 Error writing localStorage key "${key}":`, error);
      return false;
    }
  }

  safeRemoveItem(key) {
    try {
      localStorage.removeItem(key);
      return true;
    } catch (error) {
      console.error(`📦 Error removing localStorage key "${key}":`, error);
      return false;
    }
  }
}

// Create singleton instance
const storageManager = new StorageManager();

export default storageManager;

// Export utility functions
export const initializeStorage = () => storageManager.initialize();
export const getStorageHealth = () => storageManager.getStorageHealth();
export const safeGetItem = (key, defaultValue) => storageManager.safeGetItem(key, defaultValue);
export const safeSetItem = (key, value) => storageManager.safeSetItem(key, value);
export const safeRemoveItem = (key) => storageManager.safeRemoveItem(key);