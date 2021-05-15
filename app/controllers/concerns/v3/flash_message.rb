# frozen_string_literal: true

module V3::FlashMessage
  extend ActiveSupport::Concern

  included do
    before_action :store_flash_message
  end

  private

  def store_flash_message
    key = flash.keys.first
    message = {type: key.to_s, message: flash[key]} if flash[key].present?

    gon.push(flash: message.presence || {})
  end
end
