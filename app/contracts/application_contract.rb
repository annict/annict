# frozen_string_literal: true

class ApplicationContract < Dry::Validation::Contract
  module Types
    include Dry::Types()

    StrippedString = Types::String.constructor(&:strip)
  end

  TypeContainer = Dry::Schema::TypeContainer.new
  TypeContainer.register("params.stripped_string", Types::StrippedString)

  config.types = TypeContainer
  config.messages.default_locale = I18n.locale
  config.messages.load_paths += %w(
    config/locales/dry_validation.en.yml
    config/locales/dry_validation.ja.yml
  )

  register_macro(:email_format) do
    # https://github.com/K-and-R/email_validator/blob/756f4226d254713333f83f534b88d174a106eb37/lib/email_validator.rb#L13
    # https://medium.com/hackernoon/the-100-correct-way-to-validate-email-addresses-7c4818f24643
    unless %r{[^\s]@[^\s]}.match?(value)
      key.failure(:email_format)
    end
  end

  register_macro(:email_exists) do
    if User.find_by(email: value)
      key.failure(:email_exists)
    end
  end

  register_macro(:username_format) do
    unless User::USERNAME_FORMAT.match?(value)
      key.failure(:username_format)
    end

    if value.length > 20
      key.failure(:username_length)
    end

    if User.find_by(username: value)
      key.failure(:username_exists)
    end
  end
end
