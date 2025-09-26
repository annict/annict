# typed: false
# frozen_string_literal: true

module WorkHeaderLoadable
  extend ActiveSupport::Concern

  private

  def set_work_header_resources
    @work = Work.only_kept.find(params[:work_id])
    @programs = @work.programs.eager_load(:channel).merge(Channel.order(:sort_number, :id)).only_kept.in_vod.where.not(vod_title_code: "").order(:vod_title_name, :id)
  end
end
