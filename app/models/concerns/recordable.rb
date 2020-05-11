# frozen_string_literal: true

module Recordable
  extend ActiveSupport::Concern

  def update_record_body_count!(prev_resource_record, next_resource_record, field:)
    body_pair = [prev_resource_record&.body.present?, next_resource_record.body.present?]

    case body_pair
    when [false, true] then increment!(field)
    when [true, false] then decrement!(field)
    end
  end
end
