# frozen_string_literal: true

class GuestAccessToken
  def writable?
    false
  end

  def application
    @application ||= Doorkeeper::Application.official
  end

  def owner
    nil
  end
end