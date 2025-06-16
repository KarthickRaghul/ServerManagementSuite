// components/server/config/config1/components/commandexecution.tsx
import React, { useState } from 'react';
import './commandexecution.css';
import { FaTerminal, FaPlay, FaCode, FaExclamationTriangle, FaCopy, FaDownload } from 'react-icons/fa';
import ModalWrapper from './modalwrapper';
import { useCommandExecution } from '../../../../../hooks/server/useCommandExecution';
import { useNotification } from '../../../../../context/NotificationContext';

const CommandExecution: React.FC = () => {
  const [showModal, setShowModal] = useState(false);
  const [command, setCommand] = useState('');
  const [output, setOutput] = useState<string>('');
  const [hasExecuted, setHasExecuted] = useState(false);

  const { executing, error, executeCommand } = useCommandExecution();
  const { addNotification } = useNotification();

  const handleExecute = async () => {
    if (!command.trim()) return;

    try {
      const result = await executeCommand(command);
      if (result !== null) {
        setOutput(result);
        setHasExecuted(true);
        addNotification({
          title: 'Command Executed',
          message: 'Command executed successfully',
          type: 'success',
          duration: 3000
        });
      }
    } catch (err) {
      console.error('Command execution failed:', err);
      addNotification({
        title: 'Execution Failed',
        message: err instanceof Error ? err.message : 'Failed to execute command',
        type: 'error',
        duration: 5000
      });
    }
  };

  const handleCancel = () => {
    setShowModal(false);
    setCommand('');
    setOutput('');
    setHasExecuted(false);
  };

  const handleCopyOutput = () => {
    if (output) {
      navigator.clipboard.writeText(output).then(() => {
        addNotification({
          title: 'Copied',
          message: 'Command output copied to clipboard',
          type: 'info',
          duration: 2000
        });
      }).catch(() => {
        addNotification({
          title: 'Copy Failed',
          message: 'Failed to copy output to clipboard',
          type: 'error',
          duration: 3000
        });
      });
    }
  };

  const handleDownloadOutput = () => {
    if (output) {
      const blob = new Blob([output], { type: 'text/plain' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `command-output-${new Date().toISOString().slice(0, 19).replace(/:/g, '-')}.txt`;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);
      
      addNotification({
        title: 'Downloaded',
        message: 'Command output downloaded successfully',
        type: 'success',
        duration: 3000
      });
    }
  };

  const handleNewCommand = () => {
    setCommand('');
    setOutput('');
    setHasExecuted(false);
  };

  return (
    <>
      <div className="config1-cmdexec-card-container" onClick={() => setShowModal(true)}>
        <div className="config1-cmdexec-icon-wrapper">
          <FaTerminal size={20} color="white" />
        </div>
        <h3>Command Execution</h3>
        <p>Execute administrative commands remotely with secure shell access</p>
      </div>

      {showModal && (
        <ModalWrapper title="Remote Command Execution" onClose={handleCancel}>
          <div className="config1-cmdexec-modal-content">
            <div className="config1-cmdexec-warning-section">
              <FaExclamationTriangle className="config1-cmdexec-warning-icon" />
              <p className="config1-cmdexec-warning-text">
                Execute administrative commands with elevated privileges. Use caution with system-level operations.
              </p>
            </div>

            {error && (
              <div className="config1-cmdexec-error-banner">
                <p>Error: {error}</p>
              </div>
            )}
            
            <div className="config1-cmdexec-input-section">
              <label className="config1-cmdexec-label">
                <FaCode className="config1-cmdexec-label-icon" />
                Command Input
              </label>
              <div className="config1-cmdexec-textarea-wrapper">
                <textarea
                  className="config1-cmdexec-textarea"
                  placeholder="# Enter your administrative command here...
# Examples:
# systemctl status nginx
# df -h
# ps aux | grep node"
                  value={command}
                  onChange={(e) => setCommand(e.target.value.slice(0, 500))}
                  rows={6}
                  disabled={executing}
                />
                <div className="config1-cmdexec-counter">
                  {command.length}/500
                </div>
              </div>
            </div>

            {hasExecuted && (
              <div className="config1-cmdexec-output-section">
                <div className="config1-cmdexec-output-header">
                  <label className="config1-cmdexec-label">
                    <FaTerminal className="config1-cmdexec-label-icon" />
                    Command Output
                  </label>
                  <div className="config1-cmdexec-output-actions">
                    <button
                      className="config1-cmdexec-output-btn"
                      onClick={handleCopyOutput}
                      title="Copy output"
                    >
                      <FaCopy />
                    </button>
                    <button
                      className="config1-cmdexec-output-btn"
                      onClick={handleDownloadOutput}
                      title="Download output"
                    >
                      <FaDownload />
                    </button>
                  </div>
                </div>
                <div className="config1-cmdexec-output-wrapper">
                  <pre className="config1-cmdexec-output">
                    {output || 'No output returned'}
                  </pre>
                </div>
              </div>
            )}

            <div className="config1-cmdexec-actions">
              <button 
                className="config1-cmdexec-btn config1-cmdexec-btn-secondary" 
                onClick={handleCancel}
                disabled={executing}
              >
                Close
              </button>
              {hasExecuted && (
                <button 
                  className="config1-cmdexec-btn config1-cmdexec-btn-info" 
                  onClick={handleNewCommand}
                  disabled={executing}
                >
                  New Command
                </button>
              )}
              <button 
                className="config1-cmdexec-btn config1-cmdexec-btn-primary" 
                onClick={handleExecute}
                disabled={!command.trim() || executing}
              >
                <FaPlay className="config1-cmdexec-btn-icon" />
                {executing ? 'Executing...' : 'Execute Command'}
              </button>
            </div>
          </div>
        </ModalWrapper>
      )}
    </>
  );
};

export default CommandExecution;
