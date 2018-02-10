# frozen_string_literal: true

module HeadersForFastly
  extend ActiveSupport::Concern

  included do
    after_action :set_vary_header
  end

  private

  # X-UA-Device: pc | mobile
  # X-Visitor-Type: user | guest
  def set_vary_header
    response.set_header("Vary", "Accept-Encoding, X-UA-Device, X-Visitor-Type")
  end
end
