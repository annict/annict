# frozen_string_literal: true

Delayed::Worker.queue_attributes = {
  low_priority: { priority: 10 }
}
