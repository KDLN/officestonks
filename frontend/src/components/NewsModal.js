import React from "react";
import "./NewsModal.css";

const NewsModal = ({ news, onClose }) => {
  if (!news) return null;
  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="news-modal" onClick={(e) => e.stopPropagation()}>
        <div className="modal-header">
          <h2>{news.title}</h2>
          <button className="close-button" onClick={onClose}>
            ×
          </button>
        </div>
        <div className="modal-body">
          <p>{news.content}</p>
        </div>
        <div className="modal-footer">
          <button className="confirm-button" onClick={onClose}>
            Close
          </button>
        </div>
      </div>
    </div>
  );
};

export default NewsModal;
