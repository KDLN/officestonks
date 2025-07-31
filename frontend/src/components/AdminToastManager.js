import React, { useState, useEffect } from 'react';
import AdminToast from './AdminToast';

const AdminToastManager = () => {
  const [toasts, setToasts] = useState([]);

  useEffect(() => {
    const handleAdminAnnouncement = (event) => {
      const { message } = event.detail;
      const newToast = {
        id: Date.now(),
        message: message
      };
      
      setToasts(currentToasts => [...currentToasts, newToast]);
    };

    window.addEventListener('adminAnnouncement', handleAdminAnnouncement);
    
    return () => {
      window.removeEventListener('adminAnnouncement', handleAdminAnnouncement);
    };
  }, []);

  const removeToast = (toastId) => {
    setToasts(currentToasts => currentToasts.filter(toast => toast.id !== toastId));
  };

  return (
    <>
      {toasts.map((toast) => (
        <AdminToast
          key={toast.id}
          message={toast.message}
          onClose={() => removeToast(toast.id)}
        />
      ))}
    </>
  );
};

export default AdminToastManager;