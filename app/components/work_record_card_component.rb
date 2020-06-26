# frozen_string_literal: true

class WorkRecordCardComponent < ApplicationComponent
  def initialize(work_entity:, work_record_entity:)
    @work_entity = work_entity
    @work_record_entity = work_record_entity
  end

  private

  attr_reader :work_entity, :work_record_entity
end
