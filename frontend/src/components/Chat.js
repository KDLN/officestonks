import React, { useState, useEffect, useRef } from 'react';
import { getRecentMessages, sendChatMessage } from '../services/chat';
import { getUserId } from '../services/auth';
import './Chat.css';

const Chat = () => {
  const [messages, setMessages] = useState([]);
  const [newMessage, setNewMessage] = useState('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [minimized, setMinimized] = useState(false);
  const messagesEndRef = useRef(null);
  const currentUserId = parseInt(getUserId());
  const messageContainerRef = useRef(null);

  const fetchMessages = async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await getRecentMessages(50);
      setMessages(data);
      setLoading(false);
    } catch (err) {
      setError('Failed to load chat messages');
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchMessages();

    // Set up WebSocket listener for new messages
    const setupWebSocketListener = () => {
      if (window.socket) {
        const removeListener = window.addListener('chat_message', (message) => {
          setMessages(prevMessages => [...prevMessages, message.data]);
        });

        // Clean up on unmount
        return removeListener;
      }
      return null;
    };

    const removeListener = setupWebSocketListener();
    
    return () => {
      if (removeListener) {
        removeListener();
      }
    };
  }, []);

  // Scroll to bottom whenever messages change
  useEffect(() => {
    if (messagesEndRef.current && !minimized) {
      messagesEndRef.current.scrollIntoView({ behavior: 'smooth' });
    }
  }, [messages, minimized]);

  const handleSubmit = async (e) => {
    e.preventDefault();
    
    if (!newMessage.trim()) {
      return;
    }
    
    try {
      await sendChatMessage(newMessage);
      setNewMessage('');
      // Note: We don't add the message here because it will come back via WebSocket
    } catch (err) {
      console.error('Failed to send message:', err);
      // Optionally show an error to the user
    }
  };

  const toggleMinimize = () => {
    setMinimized(!minimized);
  };

  const formatTime = (dateString) => {
    const date = new Date(dateString);
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  };

  return (
    <div className={`chat-container ${minimized ? 'minimized' : ''}`}>
      <div className="chat-header" onClick={toggleMinimize}>
        <h3>Office Stonks Chat</h3>
        <button className="chat-toggle-btn">
          {minimized ? '▲' : '▼'}
        </button>
      </div>
      
      <div className="chat-messages" ref={messageContainerRef}>
        {loading ? (
          <div className="chat-status">Loading messages...</div>
        ) : error ? (
          <div className="chat-status">{error}</div>
        ) : messages.length === 0 ? (
          <div className="empty-chat">
            <p>No messages yet</p>
            <p>Be the first to start the conversation!</p>
          </div>
        ) : (
          messages.map((msg) => (
            <div 
              key={msg.id} 
              className={`message ${msg.user_id === currentUserId ? 'own' : 'other'}`}
            >
              {msg.user_id !== currentUserId && (
                <span className="message-user">{msg.username}</span>
              )}
              <div className="message-content">{msg.message}</div>
              <div className="message-time">{formatTime(msg.created_at)}</div>
            </div>
          ))
        )}
        <div ref={messagesEndRef} />
      </div>
      
      <form className="chat-form" onSubmit={handleSubmit}>
        <input
          type="text"
          className="chat-input"
          value={newMessage}
          onChange={(e) => setNewMessage(e.target.value)}
          placeholder="Type a message..."
          disabled={loading}
        />
        <button 
          type="submit" 
          className="chat-send-btn"
          disabled={!newMessage.trim() || loading}
        >
          ➤
        </button>
      </form>
    </div>
  );
};

export default Chat;