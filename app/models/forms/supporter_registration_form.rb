# frozen_string_literal: true

module Forms
  class SupporterRegistrationForm < Forms::ApplicationForm
    include Supportable

    attr_writer :auth

    validates :email, presence: true
    validates :provider_name, presence: true
    validates :provider_uid, presence: true
    validates :provider_token, presence: true

    def email
      @email ||= @auth&.dig(:extra, :raw_info, :user, :email)
    end

    def provider_name
      @auth[:provider]
    end

    def provider_uid
      @auth[:uid]
    end

    def provider_token
      @auth[:credentials][:token]
    end

    def subscriber
      @subscriber ||= gumroad_client.fetch_subscriber_by_email(email)
    end
  end
end
