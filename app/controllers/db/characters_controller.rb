# frozen_string_literal: true

module Db
  class CharactersController < ApplicationController
    def index(page: nil)
      @characters = Character.order(id: :desc).page(page)
    end
  end
end
