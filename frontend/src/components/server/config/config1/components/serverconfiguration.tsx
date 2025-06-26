import React, { useState, useEffect } from "react";
import "./serverconfiguration.css";
import { FaServer, FaCog, FaGlobe } from "react-icons/fa";
import ModalWrapper from "./modalwrapper";
import { useServerConfiguration } from "../../../../../hooks";
import { useNotification } from "../../../../../context/NotificationContext";
import { useAppContext } from "../../../../../context/AppContext";

// Comprehensive timezone mapping between common names and OS-specific identifiers
const TIMEZONE_MAP: Record<
  string,
  { linux: string; windows: string }
> = {
  // UTC
  "UTC": { linux: "UTC", windows: "UTC" },
  "GMT": { linux: "GMT", windows: "GMT Standard Time" },
  
  // North America - US
  "Eastern Time": {
    linux: "America/New_York",
    windows: "Eastern Standard Time",
  },
  "Central Time": {
    linux: "America/Chicago",
    windows: "Central Standard Time",
  },
  "Mountain Time": {
    linux: "America/Denver",
    windows: "Mountain Standard Time",
  },
  "Pacific Time": {
    linux: "America/Los_Angeles",
    windows: "Pacific Standard Time",
  },
  "Alaska Time": {
    linux: "America/Anchorage",
    windows: "Alaskan Standard Time",
  },
  "Hawaii Time": {
    linux: "Pacific/Honolulu",
    windows: "Hawaiian Standard Time",
  },
  "Arizona Time": {
    linux: "America/Phoenix",
    windows: "US Mountain Standard Time",
  },
  
  // North America - Canada
  "Atlantic Time": {
    linux: "America/Halifax",
    windows: "Atlantic Standard Time",
  },
  "Newfoundland Time": {
    linux: "America/St_Johns",
    windows: "Newfoundland Standard Time",
  },
  
  // Europe
  "London Time": {
    linux: "Europe/London",
    windows: "GMT Standard Time",
  },
  "Central European Time": {
    linux: "Europe/Berlin",
    windows: "W. Europe Standard Time",
  },
  "Eastern European Time": {
    linux: "Europe/Bucharest",
    windows: "GTB Standard Time",
  },
  "Western European Time": {
    linux: "Europe/Lisbon",
    windows: "GMT Standard Time",
  },
  "Moscow Time": {
    linux: "Europe/Moscow",
    windows: "Russian Standard Time",
  },
  "Rome Time": {
    linux: "Europe/Rome",
    windows: "W. Europe Standard Time",
  },
  "Paris Time": {
    linux: "Europe/Paris",
    windows: "Romance Standard Time",
  },
  "Amsterdam Time": {
    linux: "Europe/Amsterdam",
    windows: "W. Europe Standard Time",
  },
  "Stockholm Time": {
    linux: "Europe/Stockholm",
    windows: "W. Europe Standard Time",
  },
  "Helsinki Time": {
    linux: "Europe/Helsinki",
    windows: "FLE Standard Time",
  },
  "Athens Time": {
    linux: "Europe/Athens",
    windows: "GTB Standard Time",
  },
  "Istanbul Time": {
    linux: "Europe/Istanbul",
    windows: "Turkey Standard Time",
  },
  
  // Asia
  "India Standard Time": {
    linux: "Asia/Kolkata",
    windows: "India Standard Time",
  },
  "China Standard Time": {
    linux: "Asia/Shanghai",
    windows: "China Standard Time",
  },
  "Japan Standard Time": {
    linux: "Asia/Tokyo",
    windows: "Tokyo Standard Time",
  },
  "Korea Standard Time": {
    linux: "Asia/Seoul",
    windows: "Korea Standard Time",
  },
  "Singapore Time": {
    linux: "Asia/Singapore",
    windows: "Singapore Standard Time",
  },
  "Hong Kong Time": {
    linux: "Asia/Hong_Kong",
    windows: "China Standard Time",
  },
  "Taiwan Time": {
    linux: "Asia/Taipei",
    windows: "Taipei Standard Time",
  },
  "Thailand Time": {
    linux: "Asia/Bangkok",
    windows: "SE Asia Standard Time",
  },
  "Indonesia Time": {
    linux: "Asia/Jakarta",
    windows: "SE Asia Standard Time",
  },
  "Philippines Time": {
    linux: "Asia/Manila",
    windows: "Singapore Standard Time",
  },
  "Malaysia Time": {
    linux: "Asia/Kuala_Lumpur",
    windows: "Singapore Standard Time",
  },
  "Vietnam Time": {
    linux: "Asia/Ho_Chi_Minh",
    windows: "SE Asia Standard Time",
  },
  "Pakistan Time": {
    linux: "Asia/Karachi",
    windows: "Pakistan Standard Time",
  },
  "Bangladesh Time": {
    linux: "Asia/Dhaka",
    windows: "Bangladesh Standard Time",
  },
  "Sri Lanka Time": {
    linux: "Asia/Colombo",
    windows: "Sri Lanka Standard Time",
  },
  "Nepal Time": {
    linux: "Asia/Kathmandu",
    windows: "Nepal Standard Time",
  },
  "Afghanistan Time": {
    linux: "Asia/Kabul",
    windows: "Afghanistan Standard Time",
  },
  "Iran Time": {
    linux: "Asia/Tehran",
    windows: "Iran Standard Time",
  },
  "Israel Time": {
    linux: "Asia/Jerusalem",
    windows: "Israel Standard Time",
  },
  "Saudi Arabia Time": {
    linux: "Asia/Riyadh",
    windows: "Arab Standard Time",
  },
  "UAE Time": {
    linux: "Asia/Dubai",
    windows: "Arabian Standard Time",
  },
  "Georgia Time": {
    linux: "Asia/Tbilisi",
    windows: "Georgian Standard Time",
  },
  "Armenia Time": {
    linux: "Asia/Yerevan",
    windows: "Caucasus Standard Time",
  },
  "Azerbaijan Time": {
    linux: "Asia/Baku",
    windows: "Azerbaijan Standard Time",
  },
  "Kazakhstan Time": {
    linux: "Asia/Almaty",
    windows: "Central Asia Standard Time",
  },
  "Uzbekistan Time": {
    linux: "Asia/Tashkent",
    windows: "West Asia Standard Time",
  },
  "Mongolia Time": {
    linux: "Asia/Ulaanbaatar",
    windows: "Ulaanbaatar Standard Time",
  },
  
  // Australia & New Zealand
  "Australia Eastern Time": {
    linux: "Australia/Sydney",
    windows: "AUS Eastern Standard Time",
  },
  "Australia Central Time": {
    linux: "Australia/Adelaide",
    windows: "Cen. Australia Standard Time",
  },
  "Australia Western Time": {
    linux: "Australia/Perth",
    windows: "W. Australia Standard Time",
  },
  "New Zealand Time": {
    linux: "Pacific/Auckland",
    windows: "New Zealand Standard Time",
  },
  "Tasmania Time": {
    linux: "Australia/Hobart",
    windows: "Tasmania Standard Time",
  },
  "Darwin Time": {
    linux: "Australia/Darwin",
    windows: "AUS Central Standard Time",
  },
  
  // Pacific Islands
  "Fiji Time": {
    linux: "Pacific/Fiji",
    windows: "Fiji Standard Time",
  },
  "Tonga Time": {
    linux: "Pacific/Tongatapu",
    windows: "Tonga Standard Time",
  },
  "Samoa Time": {
    linux: "Pacific/Apia",
    windows: "Samoa Standard Time",
  },
  "Guam Time": {
    linux: "Pacific/Guam",
    windows: "West Pacific Standard Time",
  },
  
  // Africa
  "South Africa Time": {
    linux: "Africa/Johannesburg",
    windows: "South Africa Standard Time",
  },
  "Egypt Time": {
    linux: "Africa/Cairo",
    windows: "Egypt Standard Time",
  },
  "Morocco Time": {
    linux: "Africa/Casablanca",
    windows: "Morocco Standard Time",
  },
  "Nigeria Time": {
    linux: "Africa/Lagos",
    windows: "W. Central Africa Standard Time",
  },
  "Ethiopia Time": {
    linux: "Africa/Addis_Ababa",
    windows: "E. Africa Standard Time",
  },
  "Kenya Time": {
    linux: "Africa/Nairobi",
    windows: "E. Africa Standard Time",
  },
  "Ghana Time": {
    linux: "Africa/Accra",
    windows: "Greenwich Standard Time",
  },
  
  // South America
  "Brazil Time": {
    linux: "America/Sao_Paulo",
    windows: "E. South America Standard Time",
  },
  "Argentina Time": {
    linux: "America/Buenos_Aires",
    windows: "Argentina Standard Time",
  },
  "Chile Time": {
    linux: "America/Santiago",
    windows: "Pacific SA Standard Time",
  },
  "Colombia Time": {
    linux: "America/Bogota",
    windows: "SA Pacific Standard Time",
  },
  "Peru Time": {
    linux: "America/Lima",
    windows: "SA Pacific Standard Time",
  },
  "Venezuela Time": {
    linux: "America/Caracas",
    windows: "Venezuela Standard Time",
  },
  "Uruguay Time": {
    linux: "America/Montevideo",
    windows: "Montevideo Standard Time",
  },
  "Paraguay Time": {
    linux: "America/Asuncion",
    windows: "Paraguay Standard Time",
  },
  "Bolivia Time": {
    linux: "America/La_Paz",
    windows: "SA Western Standard Time",
  },
  "Ecuador Time": {
    linux: "America/Guayaquil",
    windows: "SA Pacific Standard Time",
  },
  
  // Mexico & Central America
  "Mexico Central Time": {
    linux: "America/Mexico_City",
    windows: "Central Standard Time (Mexico)",
  },
  "Mexico Pacific Time": {
    linux: "America/Mazatlan",
    windows: "Mountain Standard Time (Mexico)",
  },
  "Costa Rica Time": {
    linux: "America/Costa_Rica",
    windows: "Central America Standard Time",
  },
  "Guatemala Time": {
    linux: "America/Guatemala",
    windows: "Central America Standard Time",
  },
  "Panama Time": {
    linux: "America/Panama",
    windows: "SA Pacific Standard Time",
  },
  
  // Caribbean
  "Cuba Time": {
    linux: "America/Havana",
    windows: "Cuba Standard Time",
  },
  "Jamaica Time": {
    linux: "America/Jamaica",
    windows: "SA Pacific Standard Time",
  },
  "Puerto Rico Time": {
    linux: "America/Puerto_Rico",
    windows: "SA Western Standard Time",
  },
  
  // Atlantic
  "Azores Time": {
    linux: "Atlantic/Azores",
    windows: "Azores Standard Time",
  },
  "Cape Verde Time": {
    linux: "Atlantic/Cape_Verde",
    windows: "Cape Verde Standard Time",
  },
  "Iceland Time": {
    linux: "Atlantic/Reykjavik",
    windows: "Greenwich Standard Time",
  },
};

// Helper to get OS-specific timezone
const getMappedTimezone = (common: string, os: string): string =>
  TIMEZONE_MAP[common]?.[os.toLowerCase() as "linux" | "windows"] || common;

const ServerConfiguration: React.FC = () => {
  const [showModal, setShowModal] = useState(false);
  const [formData, setFormData] = useState({
    hostname: "",
    timezone: "UTC", // default common name
  });
  const [submitError, setSubmitError] = useState<string | null>(null);

  const { data, loading, error, updating, updateConfiguration } =
    useServerConfiguration();
  const { addNotification } = useNotification();
  const { activeDevice } = useAppContext();

  useEffect(() => {
    if (data) {
      // Reverse map: convert backend OS-specific timezone to common name
      const commonTZ = Object.entries(TIMEZONE_MAP).find(
        ([, map]) => map.linux === data.timezone || map.windows === data.timezone
      )?.[0] || data.timezone;

      setFormData({
        hostname: data.hostname,
        timezone: commonTZ,
      });
    }
  }, [data]);

  const handleApply = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitError(null);

    if (formData.hostname.trim().length > 15) {
      setSubmitError("Hostname must be 15 characters or less");
      return;
    }

    if (!/^[a-zA-Z0-9-]+$/.test(formData.hostname.trim())) {
      setSubmitError("Hostname can only contain letters, numbers, and hyphens");
      return;
    }

    const os = activeDevice?.os || "linux";
    const finalTZ = getMappedTimezone(formData.timezone, os);

    try {
      const success = await updateConfiguration({
        hostname: formData.hostname.trim(),
        timezone: finalTZ,
      });

      if (success) {
        addNotification({
          title: "Configuration Updated",
          message:
            "Server hostname and timezone have been successfully updated.",
          type: "success",
          duration: 6000,
        });
        setShowModal(false);
      }
    } catch (err) {
      const msg = err instanceof Error ? err.message : "Unknown error";
      setSubmitError(msg);
      addNotification({
        title: "Update Failed",
        message: msg,
        type: "error",
        duration: 5000,
      });
    }
  };

  const handleClose = () => {
    setShowModal(false);
    setSubmitError(null);
    if (data) {
      // Use reverse mapping when closing
      const commonTZ = Object.entries(TIMEZONE_MAP).find(
        ([, map]) => map.linux === data.timezone || map.windows === data.timezone
      )?.[0] || data.timezone;
      
      setFormData({
        hostname: data.hostname,
        timezone: commonTZ,
      });
    }
  };

  const handleOpenModal = () => {
    setSubmitError(null);
    if (data) {
      // Use reverse mapping when opening
      const commonTZ = Object.entries(TIMEZONE_MAP).find(
        ([, map]) => map.linux === data.timezone || map.windows === data.timezone
      )?.[0] || data.timezone;
      
      setFormData({
        hostname: data.hostname,
        timezone: commonTZ,
      });
    }
    setShowModal(true);
  };

  // Group timezones by region for better UX
  const getGroupedTimezones = () => {
    const groups: Record<string, string[]> = {
      "UTC/GMT": [],
      "North America": [],
      "Europe": [],
      "Asia": [],
      "Australia & Pacific": [],
      "Africa": [],
      "South America": [],
      "Central America & Caribbean": [],
      "Atlantic": [],
    };

    Object.keys(TIMEZONE_MAP).forEach((tz) => {
      if (tz.includes("UTC") || tz.includes("GMT")) {
        groups["UTC/GMT"].push(tz);
      } else if (tz.includes("Eastern Time") || tz.includes("Central Time") || 
                 tz.includes("Mountain Time") || tz.includes("Pacific Time") ||
                 tz.includes("Alaska") || tz.includes("Hawaii") || tz.includes("Arizona") ||
                 tz.includes("Atlantic Time") || tz.includes("Newfoundland")) {
        groups["North America"].push(tz);
      } else if (tz.includes("London") || tz.includes("European") || tz.includes("Moscow") ||
                 tz.includes("Rome") || tz.includes("Paris") || tz.includes("Amsterdam") ||
                 tz.includes("Stockholm") || tz.includes("Helsinki") || tz.includes("Athens") ||
                 tz.includes("Istanbul")) {
        groups["Europe"].push(tz);
      } else if (tz.includes("India") || tz.includes("China") || tz.includes("Japan") ||
                 tz.includes("Korea") || tz.includes("Singapore") || tz.includes("Hong Kong") ||
                 tz.includes("Taiwan") || tz.includes("Thailand") || tz.includes("Indonesia") ||
                 tz.includes("Philippines") || tz.includes("Malaysia") || tz.includes("Vietnam") ||
                 tz.includes("Pakistan") || tz.includes("Bangladesh") || tz.includes("Sri Lanka") ||
                 tz.includes("Nepal") || tz.includes("Afghanistan") || tz.includes("Iran") ||
                 tz.includes("Israel") || tz.includes("Saudi") || tz.includes("UAE") ||
                 tz.includes("Georgia") || tz.includes("Armenia") || tz.includes("Azerbaijan") ||
                 tz.includes("Kazakhstan") || tz.includes("Uzbekistan") || tz.includes("Mongolia")) {
        groups["Asia"].push(tz);
      } else if (tz.includes("Australia") || tz.includes("New Zealand") ||
                 tz.includes("Tasmania") || tz.includes("Darwin") || tz.includes("Fiji") ||
                 tz.includes("Tonga") || tz.includes("Samoa") || tz.includes("Guam")) {
        groups["Australia & Pacific"].push(tz);
      } else if (tz.includes("Africa") || tz.includes("South Africa") || tz.includes("Egypt") ||
                 tz.includes("Morocco") || tz.includes("Nigeria") || tz.includes("Ethiopia") ||
                 tz.includes("Kenya") || tz.includes("Ghana")) {
        groups["Africa"].push(tz);
      } else if (tz.includes("Brazil") || tz.includes("Argentina") || tz.includes("Chile") ||
                 tz.includes("Colombia") || tz.includes("Peru") || tz.includes("Venezuela") ||
                 tz.includes("Uruguay") || tz.includes("Paraguay") || tz.includes("Bolivia") ||
                 tz.includes("Ecuador")) {
        groups["South America"].push(tz);
      } else if (tz.includes("Mexico") || tz.includes("Costa Rica") || tz.includes("Guatemala") ||
                 tz.includes("Panama") || tz.includes("Cuba") || tz.includes("Jamaica") ||
                 tz.includes("Puerto Rico")) {
        groups["Central America & Caribbean"].push(tz);
      } else if (tz.includes("Azores") || tz.includes("Cape Verde") || tz.includes("Iceland")) {
        groups["Atlantic"].push(tz);
      }
    });

    return groups;
  };

  const groupedTimezones = getGroupedTimezones();

  return (
    <>
      <div className="config1-serverconfig-card-container" onClick={handleOpenModal}>
        <div className="config1-serverconfig-icon-wrapper">
          <FaServer size={20} color="white" />
        </div>
        <h3>System Configuration</h3>
        <p>Configure server hostname and timezone settings</p>
        {loading && <div className="config1-serverconfig-loading">Loading...</div>}
        {error && <div className="config1-serverconfig-error">Error loading config</div>}
      </div>

      {showModal && (
        <ModalWrapper title="System Configuration" onClose={handleClose}>
          <div className="config1-serverconfig-modal-content">
            <form className="config1-serverconfig-form" onSubmit={handleApply}>
              <div className="config1-serverconfig-input-group">
                <label className="config1-serverconfig-label">
                  <FaCog className="config1-serverconfig-label-icon" />
                  Server Hostname
                </label>
                <input
                  className="config1-serverconfig-input"
                  name="hostname"
                  placeholder="server01"
                  value={formData.hostname}
                  onChange={(e) =>
                    setFormData({ ...formData, hostname: e.target.value })
                  }
                  required
                  disabled={updating}
                  maxLength={15}
                  pattern="[a-zA-Z0-9-]+"
                />
              </div>

              <div className="config1-serverconfig-input-group">
                <label className="config1-serverconfig-label">
                  <FaGlobe className="config1-serverconfig-label-icon" />
                  System Timezone
                </label>
                <select
                  className="config1-serverconfig-select"
                  value={formData.timezone}
                  onChange={(e) =>
                    setFormData({ ...formData, timezone: e.target.value })
                  }
                  disabled={updating}
                >
                  {Object.entries(groupedTimezones).map(([group, timezones]) => (
                    timezones.length > 0 && (
                      <optgroup key={group} label={group}>
                        {timezones.map((tz) => (
                          <option key={tz} value={tz}>
                            {tz}
                          </option>
                        ))}
                      </optgroup>
                    )
                  ))}
                </select>
              </div>

              {submitError && (
                <div className="config1-serverconfig-error-banner">
                  <p>{submitError}</p>
                </div>
              )}

              <div className="config1-serverconfig-actions">
                <button
                  type="button"
                  className="config1-serverconfig-btn config1-serverconfig-btn-secondary"
                  onClick={handleClose}
                  disabled={updating}
                >
                  Close
                </button>
                <button
                  type="submit"
                  className="config1-serverconfig-btn config1-serverconfig-btn-primary"
                  disabled={!formData.hostname.trim() || updating}
                >
                  <FaServer className="config1-serverconfig-btn-icon" />
                  {updating ? "Applying..." : "Apply Configuration"}
                </button>
              </div>
            </form>
          </div>
        </ModalWrapper>
      )}
    </>
  );
};

export default ServerConfiguration;