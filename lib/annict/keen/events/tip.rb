# frozen_string_literal: true

module Annict
  module Keen
    module Events
      class Tip < Annict::Keen::Events::Application
        def close(user, slug)
          props = properties(:close, user, slug)
          ::Keen.delay(priority: 10).publish(:tips, props)
        end

        private

        def properties(action, user, slug)
          {
            action: action,
            user_id: user.encoded_id,
            device: browser.device.mobile? ? "mobile" : "pc",
            client_uuid: @request.cookies["ann_client_uuid"],
            locale: I18n.locale,
            slug: slug,
            keen: { timestamp: user.updated_at }
          }
        end
      end
    end
  end
end
