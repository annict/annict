# frozen_string_literal: true

module HeadersForFastly
  extend ActiveSupport::Concern

  included do
    after_action :set_vary_header
  end

  private

  def set_vary_header
    response.set_header("Vary", "Accept-Encoding, Origin, X-UA-Device")
  end
end
