import React, { useEffect, useState } from 'react';
import Navigation from '../components/Navigation';
import { authenticatedFetch } from '../services/authBridge';
import './AdminPanel.css';

const AuditLog = () => {
  const [events, setEvents] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchEvents = async () => {
      try {
        const res = await authenticatedFetch('/api/admin/audit?limit=100');
        if (!res.ok) throw new Error('Failed to fetch audit log');
        const data = await res.json();
        setEvents(Array.isArray(data) ? data : []);
      } catch (err) {
        console.error('Audit log fetch error:', err);
      } finally {
        setLoading(false);
      }
    };
    fetchEvents();
  }, []);

  return (
    <div className="admin-panel-page">
      <Navigation />
      <div className="admin-panel-container">
        <h1>Audit Log</h1>
        {loading ? (
          <div>Loading...</div>
        ) : (
          <table className="users-table">
            <thead>
              <tr>
                <th>Time</th>
                <th>User ID</th>
                <th>Action</th>
                <th>IP</th>
              </tr>
            </thead>
            <tbody>
              {events.map(evt => (
                <tr key={evt.id}>
                  <td>{new Date(evt.created_at).toLocaleString()}</td>
                  <td>{evt.user_id}</td>
                  <td>{evt.action}</td>
                  <td>{evt.ip_address || '-'}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
};

export default AuditLog;
