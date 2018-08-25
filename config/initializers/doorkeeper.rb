# frozen_string_literal: true

Doorkeeper.configure do
  # Change the ORM that doorkeeper will use (needs plugins)
  orm :active_record

  # This block will be called to check whether the resource owner is authenticated or not.
  resource_owner_authenticator do
    current_user || redirect_to(new_user_session_url(back: request.fullpath))
  end

  # If you want to restrict access to the web interface for adding oauth authorized
  # applications, you need to declare the block below.
  admin_authenticator do
    current_user || redirect_to(new_user_session_url(back: request.fullpath))
  end

  # Authorization Code expiration time (default 10 minutes).
  # authorization_code_expires_in 10.minutes

  # Access token expiration time (default 2 hours).
  # If you want to disable expiration, set this to nil.
  access_token_expires_in nil

  # Assign a custom TTL for implicit grants.
  # custom_access_token_expires_in do |oauth_client|
  #   oauth_client.application.additional_settings.implicit_oauth_expiration
  # end

  # Use a custom class for generating the access token.
  # https://github.com/doorkeeper-gem/doorkeeper#custom-access-token-generator
  # access_token_generator "::Doorkeeper::JWT"

  # Reuse access token for the same resource owner within an application
  # (disabled by default)
  # Rationale: https://github.com/doorkeeper-gem/doorkeeper/issues/383
  # reuse_access_token

  # Issue access tokens with refresh token (disabled by default)
  # use_refresh_token

  # Provide support for an owner to be assigned to each registered application
  # (disabled by default)
  # Optional parameter confirmation: true (default false) if you want to enforce
  # ownership of a registered application
  # Note: you must also run the rails g doorkeeper:application_owner generator
  # to provide the necessary support
  enable_application_owner confirmation: false

  # Define access token scopes for your provider
  # For more information go to
  # https://github.com/doorkeeper-gem/doorkeeper/wiki/Using-Scopes
  default_scopes  :read
  optional_scopes :write

  # Change the way client credentials are retrieved from the request object.
  # By default it retrieves first from the `HTTP_AUTHORIZATION` header, then
  # falls back to the `:client_id` and `:client_secret` params from the `params` object.
  # Check out the wiki for more information on customization
  # client_credentials :from_basic, :from_params

  # Change the way access token is authenticated from the request object.
  # By default it retrieves first from the `HTTP_AUTHORIZATION` header, then
  # falls back to the `:access_token` or `:bearer_token` params from the `params` object.
  # Check out the wiki for more information on customization
  # access_token_methods :from_bearer_authorization,
  #   :from_access_token_param,
  #   :from_bearer_param

  # Change the native redirect uri for client apps
  # When clients register with the following redirect uri,
  # they won't be redirected to any server and the authorization code will be
  # displayed within the provider
  # The value can be any string. Use nil to disable this feature.
  # When disabled, clients must provide a valid URL
  # (Similar behaviour: https://developers.google.com/accounts/docs/OAuth2InstalledApp#choosingredirecturi)
  #
  # native_redirect_uri 'urn:ietf:wg:oauth:2.0:oob'

  # Forces the usage of the HTTPS protocol in non-native redirect uris (enabled
  # by default in non-development environments). OAuth2 delegates security in
  # communication to the HTTPS protocol so it is wise to keep this enabled.
  #
  force_ssl_in_redirect_uri false

  # Specify what grant flows are enabled in array of Strings. The valid
  # strings and the flows they enable are:
  #
  # "authorization_code" => Authorization Code Grant Flow
  # "implicit"           => Implicit Grant Flow
  # "password"           => Resource Owner Password Credentials Grant Flow
  # "client_credentials" => Client Credentials Grant Flow
  #
  # If not specified, Doorkeeper enables authorization_code and
  # client_credentials.
  #
  # implicit and password grant flows have risks that you should understand
  # before enabling:
  #   http://tools.ietf.org/html/rfc6819#section-4.4.2
  #   http://tools.ietf.org/html/rfc6819#section-4.4.3
  #
  # grant_flows %w(authorization_code client_credentials)

  # Under some circumstances you might want to have applications auto-approved,
  # so that the user skips the authorization step.
  # For example if dealing with a trusted application.
  # skip_authorization do |resource_owner, client|
  #   client.superapp? or resource_owner.admin?
  # end

  # WWW-Authenticate Realm (default "Doorkeeper").
  # realm "Doorkeeper"

  base_controller "Oauth::ApplicationController"
end

Doorkeeper::Application.class_eval do
  include AASM

  aasm do
    state :published, initial: true
    state :hidden

    event :hide do
      transitions from: :published, to: :hidden
    end
  end

  scope :available, -> { published.where.not(owner: nil) }
  scope :unavailable, -> {
    unscoped.where(aasm_state: ["hidden"]).or(where(owner: nil))
  }
  scope :authorized, -> { where(oauth_access_tokens: { revoked_at: nil }) }

  def self.official
    find_by(uid: ENV.fetch("ANNICT_OFFICIAL_OAUTH_APP_UID"))
  end

  def official?
    uid == ENV.fetch("ANNICT_OFFICIAL_OAUTH_APP_UID")
  end
end

Doorkeeper::AccessToken.class_eval do
  belongs_to :user, foreign_key: :resource_owner_id, optional: true
  belongs_to :guest, optional: true

  scope :available, -> { where(revoked_at: nil) }
  scope :personal, -> { where(application_id: nil) }

  validates :description, presence: { on: :personal }

  before_validation :generate_token, on: %i(create personal)

  def owner
    user.presence || guest
  end

  def writable?
    scopes.include?("write")
  end
end

# Copy from lib/doorkeeper/oauth/helpers/uri_checker.rb to allow a fragment URI
# https://github.com/doorkeeper-gem/doorkeeper/blob/c846335ad75f2f9c7108e577cb84eaf8ab66b86f/lib/doorkeeper/oauth/helpers/uri_checker.rb
module Doorkeeper
  module OAuth
    module Helpers
      module URIChecker
        def self.valid?(url)
          # Return true if the url is native URI (urn:ietf:wg:oauth:2.0:oob).
          # This patch would be able to remove when #1060 will be released.
          # https://github.com/doorkeeper-gem/doorkeeper/pull/1060
          return true if native_uri?(url)

          uri = as_uri(url)
          !uri.host.nil? && !uri.scheme.nil?
        rescue URI::InvalidURIError
          false
        end
      end
    end
  end
end
