# frozen_string_literal: true

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
