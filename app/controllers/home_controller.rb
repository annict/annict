# frozen_string_literal: true

class HomeController < ApplicationController
  def index
    render user_signed_in? ? "index" : "index_guest"
  end
end
