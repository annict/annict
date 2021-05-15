# frozen_string_literal: true

module Annict::V4::Graphql
  class InternalClient
    def initialize(viewer:)
      @viewer = viewer
    end

    def execute(definition, variables: {})
      Canary::AnnictSchema.execute(definition, variables: variables, context: context)
    end

    private

    attr_reader :viewer

    def context
      {
        writable: true,
        admin: true,
        viewer: viewer
      }
    end
  end
end
