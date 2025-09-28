# 🏷️ Admin Frontend Barcode Implementation Guide

## 📋 Overview

This guide provides comprehensive instructions for implementing the warranty barcode generator and list functionality in the SmartSeller admin frontend. All API endpoints have been tested and are fully functional.

**Target Audience**: Frontend developers implementing the admin dashboard  
**Backend Status**: ✅ All endpoints tested and operational  
**Authentication**: JWT Bearer token required for all endpoints

---

## 🔐 Authentication Setup

### Prerequisites
- Admin user authentication system
- JWT token management
- HTTP client configuration (Axios/Fetch)

### Authentication Flow
```javascript
// Login to get access token
const loginResponse = await fetch('/api/v1/auth/login', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    email_or_phone: 'admin@example.com',
    password: 'your_password'
  })
});

const { access_token } = await loginResponse.json();

// Use token for subsequent requests
const headers = {
  'Authorization': `Bearer ${access_token}`,
  'Content-Type': 'application/json'
};
```

---

## 🎯 Core Features to Implement

### 1. Barcode Generator Interface
### 2. Barcode List & Management
### 3. Barcode Statistics Dashboard
### 4. Barcode Validation Tool

---

## 🏗️ Implementation Details

## 1. 📝 Barcode Generator Component

### API Endpoint
```
POST /api/v1/admin/warranty/barcodes/generate
```

### Required Form Fields
```typescript
interface BarcodeGenerationForm {
  product_id: string;        // UUID format required
  quantity: number;          // 1-1000
  batch_name?: string;       // Optional, max 100 chars
  notes?: string;           // Optional, max 500 chars
  expiry_months: number;    // 1-120 months
}
```

### React Component Example
```jsx
import React, { useState } from 'react';

const BarcodeGenerator = () => {
  const [formData, setFormData] = useState({
    product_id: '',
    quantity: 1,
    batch_name: '',
    notes: '',
    expiry_months: 12
  });
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState(null);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    
    try {
      const response = await fetch('/api/v1/admin/warranty/barcodes/generate', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${getAccessToken()}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(formData)
      });
      
      const data = await response.json();
      
      if (response.ok) {
        setResult(data.data);
        // Show success message
        alert(`Successfully generated ${data.data.success_count} barcodes!`);
      } else {
        // Handle validation errors
        console.error('Generation failed:', data);
      }
    } catch (error) {
      console.error('Network error:', error);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="barcode-generator">
      <h2>Generate Warranty Barcodes</h2>
      
      <form onSubmit={handleSubmit}>
        {/* Product Selection */}
        <div className="form-group">
          <label>Product *</label>
          <select 
            value={formData.product_id}
            onChange={(e) => setFormData({...formData, product_id: e.target.value})}
            required
          >
            <option value="">Select Product</option>
            {/* Populate from products API */}
          </select>
        </div>

        {/* Quantity */}
        <div className="form-group">
          <label>Quantity *</label>
          <input
            type="number"
            min="1"
            max="1000"
            value={formData.quantity}
            onChange={(e) => setFormData({...formData, quantity: parseInt(e.target.value)})}
            required
          />
        </div>

        {/* Expiry Period */}
        <div className="form-group">
          <label>Warranty Period (Months) *</label>
          <input
            type="number"
            min="1"
            max="120"
            value={formData.expiry_months}
            onChange={(e) => setFormData({...formData, expiry_months: parseInt(e.target.value)})}
            required
          />
        </div>

        {/* Batch Name */}
        <div className="form-group">
          <label>Batch Name</label>
          <input
            type="text"
            maxLength="100"
            value={formData.batch_name}
            onChange={(e) => setFormData({...formData, batch_name: e.target.value})}
            placeholder="e.g., Production Batch #001"
          />
        </div>

        {/* Notes */}
        <div className="form-group">
          <label>Notes</label>
          <textarea
            maxLength="500"
            value={formData.notes}
            onChange={(e) => setFormData({...formData, notes: e.target.value})}
            placeholder="Additional notes about this batch..."
          />
        </div>

        <button type="submit" disabled={loading}>
          {loading ? 'Generating...' : 'Generate Barcodes'}
        </button>
      </form>

      {/* Results Display */}
      {result && (
        <div className="generation-result">
          <h3>Generation Complete</h3>
          <p>Total Processed: {result.total_processed}</p>
          <p>Success Count: {result.success_count}</p>
          <p>Failure Count: {result.failure_count}</p>
          <p>Processing Time: {result.processing_time}</p>
        </div>
      )}
    </div>
  );
};
```

---

## 2. 📋 Barcode List Component

### API Endpoint
```
GET /api/v1/admin/warranty/barcodes/
```

### Query Parameters
```typescript
interface BarcodeListParams {
  page?: number;           // Default: 1
  limit?: number;          // Default: 20
  product_id?: string;     // Filter by product UUID
  batch_id?: string;       // Filter by batch UUID
  status?: 'active' | 'inactive' | 'claimed';
  search?: string;         // Search barcode values
  created_after?: string;  // Date filter (YYYY-MM-DD)
  created_before?: string; // Date filter (YYYY-MM-DD)
}
```

### React Component Example
```jsx
import React, { useState, useEffect } from 'react';

const BarcodeList = () => {
  const [barcodes, setBarcodes] = useState([]);
  const [pagination, setPagination] = useState({});
  const [filters, setFilters] = useState({});
  const [loading, setLoading] = useState(false);
  const [searchParams, setSearchParams] = useState({
    page: 1,
    limit: 20,
    status: '',
    search: ''
  });

  const fetchBarcodes = async () => {
    setLoading(true);
    try {
      const queryString = new URLSearchParams(
        Object.entries(searchParams).filter(([_, value]) => value)
      ).toString();
      
      const response = await fetch(
        `/api/v1/admin/warranty/barcodes/?${queryString}`,
        {
          headers: {
            'Authorization': `Bearer ${getAccessToken()}`
          }
        }
      );
      
      const data = await response.json();
      
      if (response.ok) {
        setBarcodes(data.data.data);
        setPagination(data.data.pagination);
        setFilters(data.data.filters);
      }
    } catch (error) {
      console.error('Failed to fetch barcodes:', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchBarcodes();
  }, [searchParams]);

  const handleSearch = (e) => {
    e.preventDefault();
    setSearchParams({...searchParams, page: 1});
  };

  return (
    <div className="barcode-list">
      <h2>Warranty Barcodes</h2>
      
      {/* Search & Filters */}
      <div className="filters">
        <form onSubmit={handleSearch}>
          <input
            type="text"
            placeholder="Search barcodes..."
            value={searchParams.search}
            onChange={(e) => setSearchParams({...searchParams, search: e.target.value})}
          />
          
          <select
            value={searchParams.status}
            onChange={(e) => setSearchParams({...searchParams, status: e.target.value})}
          >
            <option value="">All Statuses</option>
            <option value="active">Active</option>
            <option value="inactive">Inactive</option>
            <option value="claimed">Claimed</option>
          </select>
          
          <button type="submit">Search</button>
        </form>
      </div>

      {/* Barcode Table */}
      <div className="barcode-table">
        {loading ? (
          <div>Loading...</div>
        ) : (
          <table>
            <thead>
              <tr>
                <th>Barcode Value</th>
                <th>Product</th>
                <th>Status</th>
                <th>Batch</th>
                <th>Expiry Date</th>
                <th>Created</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {barcodes.map((barcode) => (
                <tr key={barcode.id}>
                  <td>
                    <code>{barcode.barcode_value}</code>
                  </td>
                  <td>
                    <div>
                      <strong>{barcode.product_name}</strong>
                      <br />
                      <small>{barcode.product_sku}</small>
                    </div>
                  </td>
                  <td>
                    <span className={`status ${barcode.status}`}>
                      {barcode.status}
                    </span>
                  </td>
                  <td>{barcode.batch_name || '-'}</td>
                  <td>{new Date(barcode.expiry_date).toLocaleDateString()}</td>
                  <td>{new Date(barcode.created_at).toLocaleDateString()}</td>
                  <td>
                    <button onClick={() => viewDetails(barcode.id)}>
                      View
                    </button>
                    {barcode.status === 'inactive' && (
                      <button onClick={() => activateBarcode(barcode.id)}>
                        Activate
                      </button>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      {/* Pagination */}
      <div className="pagination">
        <button 
          disabled={!pagination.has_prev}
          onClick={() => setSearchParams({...searchParams, page: searchParams.page - 1})}
        >
          Previous
        </button>
        
        <span>
          Page {pagination.page} of {pagination.total_pages}
        </span>
        
        <button 
          disabled={!pagination.has_next}
          onClick={() => setSearchParams({...searchParams, page: searchParams.page + 1})}
        >
          Next
        </button>
      </div>
    </div>
  );
};
```

---

## 3. 📊 Barcode Statistics Dashboard

### API Endpoint
```
GET /api/v1/admin/warranty/barcodes/stats
```

### React Component Example
```jsx
import React, { useState, useEffect } from 'react';

const BarcodeStatsDashboard = () => {
  const [stats, setStats] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchStats = async () => {
      try {
        const response = await fetch('/api/v1/admin/warranty/barcodes/stats', {
          headers: {
            'Authorization': `Bearer ${getAccessToken()}`
          }
        });
        
        const data = await response.json();
        if (response.ok) {
          setStats(data.data);
        }
      } catch (error) {
        console.error('Failed to fetch stats:', error);
      } finally {
        setLoading(false);
      }
    };

    fetchStats();
  }, []);

  if (loading) return <div>Loading statistics...</div>;
  if (!stats) return <div>Failed to load statistics</div>;

  return (
    <div className="barcode-stats-dashboard">
      <h2>Barcode Statistics</h2>
      
      {/* Overview Cards */}
      <div className="stats-grid">
        <div className="stat-card">
          <h3>Total Barcodes</h3>
          <div className="stat-value">{stats.total_barcodes.toLocaleString()}</div>
        </div>
        
        <div className="stat-card">
          <h3>Active Barcodes</h3>
          <div className="stat-value">{stats.active_barcodes.toLocaleString()}</div>
        </div>
        
        <div className="stat-card">
          <h3>Claimed Barcodes</h3>
          <div className="stat-value">{stats.claimed_barcodes.toLocaleString()}</div>
        </div>
        
        <div className="stat-card">
          <h3>Expired Barcodes</h3>
          <div className="stat-value">{stats.expired_barcodes.toLocaleString()}</div>
        </div>
      </div>

      {/* Recent Activity */}
      <div className="recent-activity">
        <h3>Recent Activity</h3>
        <div className="activity-grid">
          <div className="activity-item">
            <span>Generated Today</span>
            <strong>{stats.generated_today}</strong>
          </div>
          <div className="activity-item">
            <span>Generated This Week</span>
            <strong>{stats.generated_this_week}</strong>
          </div>
          <div className="activity-item">
            <span>Generated This Month</span>
            <strong>{stats.generated_this_month}</strong>
          </div>
        </div>
      </div>

      {/* Status Breakdown */}
      <div className="status-breakdown">
        <h3>Status Breakdown</h3>
        <div className="breakdown-chart">
          {Object.entries(stats.status_breakdown).map(([status, count]) => (
            <div key={status} className="breakdown-item">
              <span className={`status-indicator ${status}`}></span>
              <span>{status}: {count}</span>
            </div>
          ))}
        </div>
      </div>
      
      <div className="last-updated">
        Last updated: {new Date(stats.last_updated).toLocaleString()}
      </div>
    </div>
  );
};
```

---

## 4. 🔍 Barcode Validation Tool

### API Endpoint
```
GET /api/v1/admin/warranty/barcodes/validate/{barcode_value}
```

### React Component Example
```jsx
import React, { useState } from 'react';

const BarcodeValidator = () => {
  const [barcodeValue, setBarcodeValue] = useState('');
  const [validationResult, setValidationResult] = useState(null);
  const [loading, setLoading] = useState(false);

  const validateBarcode = async (e) => {
    e.preventDefault();
    if (!barcodeValue.trim()) return;
    
    setLoading(true);
    try {
      const response = await fetch(
        `/api/v1/admin/warranty/barcodes/validate/${encodeURIComponent(barcodeValue)}`,
        {
          headers: {
            'Authorization': `Bearer ${getAccessToken()}`
          }
        }
      );
      
      const data = await response.json();
      setValidationResult(data.data);
    } catch (error) {
      console.error('Validation failed:', error);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="barcode-validator">
      <h2>Validate Barcode</h2>
      
      <form onSubmit={validateBarcode}>
        <div className="form-group">
          <label>Barcode Value</label>
          <input
            type="text"
            value={barcodeValue}
            onChange={(e) => setBarcodeValue(e.target.value)}
            placeholder="Enter barcode value (e.g., REX24ABC123DEF456)"
            required
          />
        </div>
        
        <button type="submit" disabled={loading}>
          {loading ? 'Validating...' : 'Validate Barcode'}
        </button>
      </form>

      {/* Validation Results */}
      {validationResult && (
        <div className={`validation-result ${validationResult.is_valid ? 'valid' : 'invalid'}`}>
          <h3>Validation Result</h3>
          
          <div className="result-grid">
            <div className="result-item">
              <strong>Status:</strong>
              <span className={`status ${validationResult.is_valid ? 'valid' : 'invalid'}`}>
                {validationResult.is_valid ? 'Valid' : 'Invalid'}
              </span>
            </div>
            
            <div className="result-item">
              <strong>Barcode:</strong>
              <code>{validationResult.barcode_value}</code>
            </div>
            
            <div className="result-item">
              <strong>Current Status:</strong>
              <span className={`status ${validationResult.status}`}>
                {validationResult.status}
              </span>
            </div>
            
            {validationResult.product && (
              <div className="result-item">
                <strong>Product:</strong>
                <div>
                  {validationResult.product.name}
                  <br />
                  <small>SKU: {validationResult.product.sku}</small>
                </div>
              </div>
            )}
            
            {validationResult.expiry_date && (
              <div className="result-item">
                <strong>Expiry Date:</strong>
                <span className={validationResult.is_expired ? 'expired' : 'valid'}>
                  {new Date(validationResult.expiry_date).toLocaleDateString()}
                  {validationResult.is_expired && ' (Expired)'}
                </span>
              </div>
            )}
            
            {validationResult.claimed_at && (
              <div className="result-item">
                <strong>Claimed:</strong>
                <span>{new Date(validationResult.claimed_at).toLocaleDateString()}</span>
              </div>
            )}
          </div>
          
          {validationResult.validation_error && (
            <div className="validation-error">
              <strong>Error:</strong> {validationResult.validation_error.message}
            </div>
          )}
        </div>
      )}
    </div>
  );
};
```

---

## 5. 📄 Barcode Details Modal

### API Endpoint
```
GET /api/v1/admin/warranty/barcodes/{id}
```

### React Component Example
```jsx
import React, { useState, useEffect } from 'react';

const BarcodeDetailsModal = ({ barcodeId, onClose }) => {
  const [barcode, setBarcode] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchBarcodeDetails = async () => {
      try {
        const response = await fetch(
          `/api/v1/admin/warranty/barcodes/${barcodeId}`,
          {
            headers: {
              'Authorization': `Bearer ${getAccessToken()}`
            }
          }
        );
        
        const data = await response.json();
        if (response.ok) {
          setBarcode(data.data);
        }
      } catch (error) {
        console.error('Failed to fetch barcode details:', error);
      } finally {
        setLoading(false);
      }
    };

    if (barcodeId) {
      fetchBarcodeDetails();
    }
  }, [barcodeId]);

  if (loading) return <div className="modal-loading">Loading...</div>;
  if (!barcode) return <div className="modal-error">Failed to load barcode details</div>;

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-content" onClick={(e) => e.stopPropagation()}>
        <div className="modal-header">
          <h2>Barcode Details</h2>
          <button className="close-button" onClick={onClose}>×</button>
        </div>
        
        <div className="modal-body">
          <div className="details-grid">
            <div className="detail-item">
              <label>Barcode Value</label>
              <code>{barcode.barcode_value}</code>
            </div>
            
            <div className="detail-item">
              <label>Product</label>
              <div>
                <strong>{barcode.product_name}</strong>
                <br />
                <small>SKU: {barcode.product_sku}</small>
              </div>
            </div>
            
            <div className="detail-item">
              <label>Status</label>
              <span className={`status ${barcode.status}`}>
                {barcode.status}
              </span>
            </div>
            
            <div className="detail-item">
              <label>Batch</label>
              <span>{barcode.batch_name || 'No batch'}</span>
            </div>
            
            <div className="detail-item">
              <label>Expiry Date</label>
              <span>{new Date(barcode.expiry_date).toLocaleDateString()}</span>
            </div>
            
            <div className="detail-item">
              <label>Active</label>
              <span className={barcode.is_active ? 'active' : 'inactive'}>
                {barcode.is_active ? 'Yes' : 'No'}
              </span>
            </div>
            
            <div className="detail-item">
              <label>Created</label>
              <span>{new Date(barcode.created_at).toLocaleString()}</span>
            </div>
            
            <div className="detail-item">
              <label>Last Updated</label>
              <span>{new Date(barcode.updated_at).toLocaleString()}</span>
            </div>
          </div>
        </div>
        
        <div className="modal-footer">
          {!barcode.is_active && (
            <button 
              className="btn-primary"
              onClick={() => activateBarcode(barcode.id)}
            >
              Activate Barcode
            </button>
          )}
          <button className="btn-secondary" onClick={onClose}>
            Close
          </button>
        </div>
      </div>
    </div>
  );
};
```

---

## 🎨 CSS Styling Examples

```css
/* Barcode Generator Styles */
.barcode-generator {
  max-width: 600px;
  margin: 0 auto;
  padding: 20px;
}

.form-group {
  margin-bottom: 20px;
}

.form-group label {
  display: block;
  margin-bottom: 5px;
  font-weight: bold;
}

.form-group input,
.form-group select,
.form-group textarea {
  width: 100%;
  padding: 10px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
}

/* Status Indicators */
.status {
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: bold;
  text-transform: uppercase;
}

.status.active {
  background-color: #d4edda;
  color: #155724;
}

.status.inactive {
  background-color: #f8d7da;
  color: #721c24;
}

.status.claimed {
  background-color: #fff3cd;
  color: #856404;
}

/* Statistics Dashboard */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 20px;
  margin-bottom: 30px;
}

.stat-card {
  background: white;
  padding: 20px;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
  text-align: center;
}

.stat-value {
  font-size: 2em;
  font-weight: bold;
  color: #007bff;
}

/* Barcode Table */
.barcode-table table {
  width: 100%;
  border-collapse: collapse;
  margin-top: 20px;
}

.barcode-table th,
.barcode-table td {
  padding: 12px;
  text-align: left;
  border-bottom: 1px solid #ddd;
}

.barcode-table th {
  background-color: #f8f9fa;
  font-weight: bold;
}

/* Modal Styles */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-content {
  background: white;
  border-radius: 8px;
  max-width: 600px;
  width: 90%;
  max-height: 80vh;
  overflow-y: auto;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px;
  border-bottom: 1px solid #ddd;
}

.close-button {
  background: none;
  border: none;
  font-size: 24px;
  cursor: pointer;
}
```

---

## 🔧 Utility Functions

```javascript
// Token management
const getAccessToken = () => {
  return localStorage.getItem('access_token');
};

const setAccessToken = (token) => {
  localStorage.setItem('access_token', token);
};

// API client wrapper
const apiClient = {
  async request(endpoint, options = {}) {
    const token = getAccessToken();
    const config = {
      headers: {
        'Content-Type': 'application/json',
        ...(token && { 'Authorization': `Bearer ${token}` }),
        ...options.headers
      },
      ...options
    };

    const response = await fetch(`/api/v1${endpoint}`, config);
    const data = await response.json();

    if (!response.ok) {
      throw new Error(data.message || 'API request failed');
    }

    return data;
  },

  // Barcode-specific methods
  async generateBarcodes(formData) {
    return this.request('/admin/warranty/barcodes/generate', {
      method: 'POST',
      body: JSON.stringify(formData)
    });
  },

  async getBarcodes(params = {}) {
    const queryString = new URLSearchParams(params).toString();
    return this.request(`/admin/warranty/barcodes/?${queryString}`);
  },

  async getBarcodeDetails(id) {
    return this.request(`/admin/warranty/barcodes/${id}`);
  },

  async getBarcodeStats() {
    return this.request('/admin/warranty/barcodes/stats');
  },

  async validateBarcode(barcodeValue) {
    return this.request(`/admin/warranty/barcodes/validate/${encodeURIComponent(barcodeValue)}`);
  }
};

// Form validation helpers
const validateUUID = (uuid) => {
  const uuidRegex = /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i;
  return uuidRegex.test(uuid);
};

const validateBarcodeForm = (formData) => {
  const errors = {};

  if (!formData.product_id) {
    errors.product_id = 'Product is required';
  } else if (!validateUUID(formData.product_id)) {
    errors.product_id = 'Invalid product ID format';
  }

  if (!formData.quantity || formData.quantity < 1 || formData.quantity > 1000) {
    errors.quantity = 'Quantity must be between 1 and 1000';
  }

  if (!formData.expiry_months || formData.expiry_months < 1 || formData.expiry_months > 120) {
    errors.expiry_months = 'Expiry period must be between 1 and 120 months';
  }

  if (formData.batch_name && formData.batch_name.length > 100) {
    errors.batch_name = 'Batch name must be less than 100 characters';
  }

  if (formData.notes && formData.notes.length > 500) {
    errors.notes = 'Notes must be less than 500 characters';
  }

  return errors;
};
```

---

## 📱 Responsive Design Considerations

```css
/* Mobile-first responsive design */
@media (max-width: 768px) {
  .stats-grid {
    grid-template-columns: 1fr;
  }
  
  .barcode-table {
    overflow-x: auto;
  }
  
  .modal-content {
    width: 95%;
    margin: 10px;
  }
  
  .filters {
    flex-direction: column;
  }
  
  .filters input,
  .filters select {
    margin-bottom: 10px;
  }
}

@media (max-width: 480px) {
  .barcode-generator {
    padding: 10px;
  }
  
  .stat-card {
    padding: 15px;
  }
  
  .stat-value {
    font-size: 1.5em;
  }
}
```

---

## 🚀 Implementation Checklist

### Phase 1: Basic Setup
- [ ] Set up authentication system
- [ ] Configure API client
- [ ] Implement token management
- [ ] Create base layout components

### Phase 2: Core Features
- [ ] Implement barcode generator form
- [ ] Create barcode list with pagination
- [ ] Add search and filtering
- [ ] Build statistics dashboard

### Phase 3: Advanced Features
- [ ] Add barcode validation tool
- [ ] Implement barcode details modal
- [ ] Create bulk operations
- [ ] Add export functionality

### Phase 4: Polish & Testing
- [ ] Add loading states
- [ ] Implement error handling
- [ ] Add responsive design
- [ ] Write unit tests
- [ ] Perform integration testing

---

## 🐛 Error Handling

```javascript
// Error handling utility
const handleApiError = (error, context = '') => {
  console.error(`API Error ${context}:`, error);
  
  if (error.message.includes('401')) {
    // Token expired, redirect to login
    localStorage.removeItem('access_token');
    window.location.href = '/login';
    return;
  }
  
  if (error.message.includes('validation')) {
    // Show validation errors to user
    showValidationErrors(error.validation_errors);
    return;
  }
  
  // Generic error message
  showErrorMessage('An unexpected error occurred. Please try again.');
};

const showErrorMessage = (message) => {
  // Implement your preferred notification system
  alert(message); // Replace with toast/notification
};

const showValidationErrors = (errors) => {
  // Display validation errors in form
  errors.forEach(error => {
    console.error('Validation error:', error);
  });
};
```

---

## 📋 Testing Guidelines

### Unit Tests
```javascript
// Example test for barcode validation
describe('Barcode Validation', () => {
  test('should validate UUID format', () => {
    expect(validateUUID('550e8400-e29b-41d4-a716-446655440000')).toBe(true);
    expect(validateUUID('invalid-uuid')).toBe(false);
  });
  
  test('should validate form data', () => {
    const validForm = {
      product_id: '550e8400-e29b-41d4-a716-446655440000',
      quantity: 10,
      expiry_months: 12
    };
    
    expect(Object.keys(validateBarcodeForm(validForm))).toHaveLength(0);
  });
});
```

### Integration Tests
- Test API endpoints with mock data
- Verify authentication flow
- Test error handling scenarios
- Validate responsive design

---

## 🔗 Related Documentation

- [Warranty Admin Endpoints OpenAPI Spec](../api/openapi/warranty-admin-endpoints.yaml)
- [Authentication Guide](./AUTH_IMPLEMENTATION_GUIDE.md)
- [Backend API Testing Results](./WARRANTY_IMPLEMENTATION_TRACKER.md)

---

## 📞 Support

For technical questions or issues:
- Check the API documentation in `/api/openapi/`
- Review backend implementation in `/internal/interfaces/api/handler/`
- All endpoints have been tested and are operational

**Last Updated**: December 29, 2024  
**API Version**: v1  
**Status**: ✅ Ready for Implementation