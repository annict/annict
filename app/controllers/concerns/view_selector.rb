# frozen_string_literal: true

module ViewSelector
  extend ActiveSupport::Concern

  included do
    before_action :register_mobile_variant
  end

  private

  def register_mobile_variant
    request.variant = :mobile if browser.device.mobile?
  end
end
