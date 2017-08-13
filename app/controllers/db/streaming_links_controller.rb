# frozen_string_literal: true

module Db
  class StreamingLinksController < Db::ApplicationController
    permits :channel_id, :locale, :unique_id

    before_action :authenticate_user!
    before_action :load_work, only: %i(index new create)
    before_action :load_streaming_link, only: %i(edit update hide destroy activities)

    def index
      @streaming_links = @work.streaming_links.order(id: :desc)
    end

    def new
      @form = DB::StreamingLinkRowsForm.new
      authorize @form, :new?
    end

    def create(db_streaming_link_rows_form)
      @form = DB::StreamingLinkRowsForm.new(db_streaming_link_rows_form.permit(:rows))
      @form.user = current_user
      @form.work = @work
      authorize @form, :create?

      return render(:new) unless @form.valid?
      @form.save!

      flash[:notice] = t("messages._common.created")
      redirect_to db_work_streaming_links_path(@work)
    end

    def edit
      authorize @streaming_link, :edit?
      @work = @streaming_link.work
    end

    def update(streaming_link)
      authorize @streaming_link, :update?
      @work = @streaming_link.work

      @streaming_link.attributes = streaming_link
      @streaming_link.user = current_user

      return render(:edit) unless @streaming_link.valid?
      @streaming_link.save_and_create_activity!

      flash[:notice] = t("messages._common.updated")
      redirect_to db_work_streaming_links_path(@work)
    end

    def hide
      authorize @streaming_link, :hide?

      @streaming_link.hide!

      flash[:notice] = t("resources.streaming_link.unpublished")
      redirect_back fallback_location: db_works_path
    end

    def destroy
      authorize @streaming_link, :destroy?

      @streaming_link.destroy

      flash[:notice] = t("resources.streaming_link.deleted")
      redirect_back fallback_location: db_works_path
    end

    def activities
      @activities = @streaming_link.db_activities.order(id: :desc)
      @comment = @streaming_link.db_comments.new
    end

    private

    def load_streaming_link
      @streaming_link = StreamingLink.find(params[:id])
    end
  end
end
