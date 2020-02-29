# frozen_string_literal: true

module ActionDispatch
  module Routing
    class Mapper
      def draw(routes_name)
        instance_eval(File.read(Rails.root.join("config/routes/#{routes_name}.rb")))
      end
    end
  end
end

class MemberConstraint
  def matches?(request)
    request.session["warden.user.user.key"].present?
  end
end

class GuestConstraint
  def matches?(request)
    !MemberConstraint.new.matches?(request)
  end
end

Rails.application.routes.draw do
  draw :api
  draw :chat
  draw :db
  draw :forum
  draw :internal_api
  draw :userland
  draw :web
end
