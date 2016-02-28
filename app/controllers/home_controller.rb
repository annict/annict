# frozen_string_literal: true

class HomeController < ApplicationController
  def index
    @works_count = Work.published.count
    render user_signed_in? ? "index" : "index_guest"
  end
end
