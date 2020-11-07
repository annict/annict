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
end
