#!/bin/bash

# mockserver_manager.sh
# This script manages the MockServer for testing external APIs
# Usage: ./mockserver_manager.sh [start|stop|restart|reset|status]

MOCKSERVER_DIR="/Users/tes/go/src/github.com/kirimku/kirimku-backend/internal/tests/mockserver"
MOCKSERVER_PORT=1080
MOCKSERVER_URL="http://localhost:${MOCKSERVER_PORT}"
EXPECTATIONS_FILE="${MOCKSERVER_DIR}/expectations.json"

function check_status() {
  echo "Checking MockServer status..."
  if curl -s "${MOCKSERVER_URL}/mockserver/status" > /dev/null; then
    echo "✅ MockServer is running at ${MOCKSERVER_URL}"
    return 0
  else
    echo "❌ MockServer is not running"
    return 1
  fi
}

function start_mockserver() {
  echo "Starting MockServer..."
  
  # Check if already running
  if check_status > /dev/null; then
    echo "MockServer is already running at ${MOCKSERVER_URL}"
    return 0
  fi
  
  # Start the MockServer
  cd "${MOCKSERVER_DIR}" && docker-compose up -d
  
  # Wait for MockServer to be ready
  echo "⏳ Waiting for MockServer to be ready..."
  for i in {1..10}; do
    if curl -s "${MOCKSERVER_URL}/mockserver/status" > /dev/null; then
      echo "✅ MockServer is ready! Running on ${MOCKSERVER_URL}"
      load_expectations
      return 0
    else
      echo "⏳ Still starting... (attempt $i/10)"
      sleep 2
    fi
  done
  
  echo "❌ Failed to start MockServer"
  return 1
}

function stop_mockserver() {
  echo "Stopping MockServer..."
  cd "${MOCKSERVER_DIR}" && docker-compose down
  echo "✅ MockServer stopped"
}

function restart_mockserver() {
  stop_mockserver
  start_mockserver
}

function reset_expectations() {
  echo "Resetting MockServer expectations..."
  if ! check_status > /dev/null; then
    echo "❌ MockServer is not running. Start it first."
    return 1
  fi
  
  curl -X PUT "${MOCKSERVER_URL}/mockserver/reset"
  echo "✅ Expectations reset"
  
  # Reload expectations if file exists
  load_expectations
}

function load_expectations() {
  # Load expectations from file if it exists
  if [ -f "$EXPECTATIONS_FILE" ]; then
    echo "Loading expectations from ${EXPECTATIONS_FILE}..."
    curl -X PUT "${MOCKSERVER_URL}/mockserver/expectation" -d @"${EXPECTATIONS_FILE}"
    echo "✅ Expectations loaded"
  else
    # Set up JNT expectations
    setup_jnt_expectations
  fi
}

function setup_jnt_expectations() {
  echo "Setting up JNT expectations..."
  
  # JNT Tariff Inquiry
  curl -s -X PUT "${MOCKSERVER_URL}/mockserver/expectation" -d '{
    "httpRequest": {
      "path": "/jandt_track/inquiry.action"
    },
    "httpResponse": {
      "statusCode": 200,
      "headers": {
        "Content-Type": ["application/json"]
      },
      "body": {
        "is_success": true,
        "content": {
          "tariff": 20000,
          "shipping_fee": 20000,
          "insurance_fee": 0,
          "options": {
            "service_code": "REG",
            "service_name": "Regular"
          }
        },
        "message": ""
      }
    }
  }' > /dev/null
  
  # JNT Booking
  curl -s -X PUT "${MOCKSERVER_URL}/mockserver/expectation" -d '{
    "httpRequest": {
      "path": "/jandt_track/order.action"
    },
    "httpResponse": {
      "statusCode": 200,
      "headers": {
        "Content-Type": ["application/json"]
      },
      "body": {
        "is_success": true,
        "content": {
          "booking_code": "JNT12345678",
          "awb": "JNT12345678",
          "status": "success",
          "shipping_cost": 20000
        },
        "message": ""
      }
    }
  }' > /dev/null
  
  # JNT Cancellation
  curl -s -X PUT "${MOCKSERVER_URL}/mockserver/expectation" -d '{
    "httpRequest": {
      "path": "/jandt_track/update.action"
    },
    "httpResponse": {
      "statusCode": 200,
      "headers": {
        "Content-Type": ["application/json"]
      },
      "body": {
        "is_success": true,
        "content": {
          "status": "success",
          "message": "Booking cancelled successfully"
        },
        "message": ""
      }
    }
  }' > /dev/null
  
  echo "✅ JNT expectations set up"
}

# Main execution
case "$1" in
  start)
    start_mockserver
    ;;
  stop)
    stop_mockserver
    ;;
  restart)
    restart_mockserver
    ;;
  reset)
    reset_expectations
    ;;
  status)
    check_status
    ;;
  *)
    echo "Usage: $0 [start|stop|restart|reset|status]"
    exit 1
    ;;
esac

exit 0