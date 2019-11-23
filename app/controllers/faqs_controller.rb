# frozen_string_literal: true

class FaqsController < ApplicationController
  def index
    @faq_categories = FaqCategory.without_deleted.order(:sort_number)
  end
end
