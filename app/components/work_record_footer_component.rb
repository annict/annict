# frozen_string_literal: true

class WorkRecordFooterComponent < ApplicationComponent
  def initialize(user_entity:, record_entity:, work_record_entity:)
    @user_entity = user_entity
    @record_entity = record_entity
    @work_record_entity = work_record_entity
  end

  private

  attr_reader :record_entity, :user_entity, :work_record_entity
end
