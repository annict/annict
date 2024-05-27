# typed: false
# frozen_string_literal: true

module Annict
  module RoutingConstraints
    class Member
      def matches?(request)
        request.session["warden.user.user.key"].present?
      end
    end

    class Guest
      def matches?(request)
        !Annict::RoutingConstraints::Member.new.matches?(request)
      end
    end
  end
end
