# frozen_string_literal: true

module LogrageSetting
  extend ActiveSupport::Concern

  included do
    def append_info_to_payload(payload)
      super
      payload[:request_id] = request.uuid
      payload[:client_uuid] = viewer_uuid
      payload[:user_id] = current_user.id if user_signed_in?
    end
  end
end
