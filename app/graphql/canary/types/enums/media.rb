# typed: false
# frozen_string_literal: true

module Canary
  module Types
    module Enums
      class Media < Canary::Types::Enums::Base
        description "メディア"

        value "TV", "テレビ"
        value "OVA", "OVA"
        value "MOVIE", "映画"
        value "WEB", "Web"
        value "OTHER", "その他"
      end
    end
  end
end
