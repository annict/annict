# frozen_string_literal: true

module HeadersForFastly
  private

  # X-Visitor-Type: user | guest
  def set_vary_header
    response.set_header("Vary", "Accept-Encoding, X-Visitor-Type")
  end
end
