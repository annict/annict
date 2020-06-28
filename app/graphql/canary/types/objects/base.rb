# frozen_string_literal: true

module Canary
  module Types
    module Objects
      class Base < GraphQL::Schema::Object
        include Imgix::Rails::UrlHelper
        include ImageHelper
        include MarkdownHelper

        def database_id
          object.id
        end

        def build_order(order_by)
          return OrderProperty.new unless order_by

          OrderProperty.new(order_by[:field], order_by[:direction])
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
