# frozen_string_literal: true

class ApplicationEntity < Dry::Struct
  module Types
    include Dry.Types(default: :strict)
  end
end
