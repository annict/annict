# typed: false
# frozen_string_literal: true

Delayed::Worker.queue_attributes = {
  default: {priority: 0},
  mailers: {priority: 10},
  low: {priority: 20}
}
