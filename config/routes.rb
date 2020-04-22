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

Rails.application.routes.draw do
  draw :api
  draw :chat
  draw :db
  draw :forum
  draw :internal_api
  draw :local_api
  draw :userland
  draw :web
end
