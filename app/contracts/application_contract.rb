# frozen_string_literal: true

class ApplicationContract < Dry::Validation::Contract
  module Types
    include Dry::Types()

    StrippedString = Types::String.constructor(&:strip)
    CoercibleBoolean = Types::Bool.constructor { |value| ActiveRecord::Type::Boolean.new.serialize(value) }
  end

  TypeContainer = Dry::Schema::TypeContainer.new
  TypeContainer.register("params.stripped_string", Types::StrippedString)
  TypeContainer.register("params.coercible_boolean", Types::CoercibleBoolean)

  config.types = TypeContainer
  config.messages.default_locale = I18n.locale
  config.messages.load_paths += %w(
    config/locales/dry_validation.en.yml
    config/locales/dry_validation.ja.yml
  )

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
