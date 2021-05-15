# frozen_string_literal: true

module V3::ViewSelector
  extend ActiveSupport::Concern

  included do
    before_action :register_mobile_variant
  end

  private

  def register_mobile_variant
    request.variant = :mobile unless device_pc?
  end
end
