# frozen_string_literal: true

module Canary
  module Types
    module Objects
      class Base < GraphQL::Schema::Object
        include GraphQL::FragmentCache::Object

        include ImageHelper
        include MarkdownHelper

        def database_id
          object.id
        end

        def method_missing(method_name, *arguments, &block)
          return super if method_name.blank?
          return super unless method_name.to_s.start_with?("local_")

          object.send(method_name)
        end

        def respond_to_missing?(method_name, include_private = false)
          method_name.to_s.start_with?("local_") || super
        end
      end
    end
  end
end
