module FlashMessage
  extend ActiveSupport::Concern

  included do
    before_action :store_flash_message
  end

  private

  def store_flash_message
    key = flash.keys.first
    message = { type: key.to_s, body: flash[key] } if flash[key].present?

    gon.push(flash: message.presence || {})
  end
end
