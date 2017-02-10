# frozen_string_literal: true

module Annict
  module Keen
    module Events
      class Tip < Annict::Keen::Events::Application
        def close(slug)
          ::Keen.delay(priority: 10).publish(:tips, properties(:close, slug))
        end

        private

        def properties(action, slug)
          {
            action: action,
            user_id: @user&.encoded_id,
            device: browser.device.mobile? ? "mobile" : "pc",
            client_uuid: @request.cookies["ann_client_uuid"],
            locale: @user&.locale,
            time_zone: @user&.time_zone,
            slug: slug,
            page_category: @params[:page_category],
            keen: { timestamp: @user&.updated_at }
          }
        end
      end
    end
  end
end
