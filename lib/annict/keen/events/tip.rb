# frozen_string_literal: true

module Annict
  module Keen
    module Events
      class Tip < Annict::Keen::Events::Application
        def close(slug)
          SendKeenEventJob.perform_later("tips", properties(:close, slug))
        end

        private

        def properties(action, slug)
          {
            action: action.to_s,
            user_id: @user&.encoded_id,
            device: browser.device.mobile? ? "mobile" : "pc",
            client_uuid: @request.cookies["ann_client_uuid"],
            locale: @user&.locale,
            time_zone: @user&.time_zone,
            slug: slug,
            page_category: @params[:page_category],
            keen: { timestamp: @user&.updated_at&.to_s }
          }
        end
      end
    end
  end
end
