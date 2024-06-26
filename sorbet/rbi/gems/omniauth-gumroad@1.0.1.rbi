# typed: true

# DO NOT EDIT MANUALLY
# This is an autogenerated file for types exported from the `omniauth-gumroad` gem.
# Please instead update this file by running `bin/tapioca gem omniauth-gumroad`.


# source://omniauth-gumroad//lib/omniauth-gumroad/version.rb#1
module OmniAuth
  class << self
    # source://omniauth/1.9.2/lib/omniauth.rb#118
    def config; end

    # source://omniauth/1.9.2/lib/omniauth.rb#122
    def configure; end

    # source://omniauth/1.9.2/lib/omniauth.rb#126
    def logger; end

    # source://omniauth/1.9.2/lib/omniauth.rb#130
    def mock_auth_for(provider); end

    # source://omniauth/1.9.2/lib/omniauth.rb#19
    def strategies; end
  end
end

# source://omniauth-gumroad//lib/omniauth-gumroad/version.rb#2
module OmniAuth::Gumroad; end

# source://omniauth-gumroad//lib/omniauth-gumroad/version.rb#3
OmniAuth::Gumroad::VERSION = T.let(T.unsafe(nil), String)

# source://omniauth-gumroad//lib/omniauth/strategies/gumroad.rb#4
module OmniAuth::Strategies; end

# source://omniauth-gumroad//lib/omniauth/strategies/gumroad.rb#5
class OmniAuth::Strategies::Gumroad < ::OmniAuth::Strategies::OAuth2
  # source://omniauth-gumroad//lib/omniauth/strategies/gumroad.rb#19
  def authorize_params; end

  # source://omniauth-gumroad//lib/omniauth/strategies/gumroad.rb#29
  def callback_url; end

  # source://omniauth-gumroad//lib/omniauth/strategies/gumroad.rb#49
  def raw_info; end

  # source://omniauth-gumroad//lib/omniauth/strategies/gumroad.rb#15
  def request_phase; end
end
