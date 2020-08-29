# frozen_string_literal: true

module Canary
  module Resolvers
    class Programs < Canary::Resolvers::Base
      def resolve(order_by: nil)
        order = Canary::OrderProperty.build(order_by)

        programs = object.programs.only_kept
        programs = programs.order(order.field => order.direction)

        programs
      end
    end
  end
end
