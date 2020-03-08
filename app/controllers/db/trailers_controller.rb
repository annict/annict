# frozen_string_literal: true

module DB
  class TrailersController < DB::ApplicationController
    before_action :authenticate_user!

    def index
      @work = Work.find(params[:work_id])
      @trailers = @work.trailers.order(:sort_number)
    end

    def new
      @work = Work.find(params[:work_id])
      @form = DB::TrailerRowsForm.new
      authorize @form, :new?
    end

    def create
      @work = Work.find(params[:work_id])
      @form = DB::TrailerRowsForm.new(trailer_rows_form_params)
      @form.user = current_user
      @form.work = @work
      authorize @form, :create?

      return render(:new) unless @form.valid?
      @form.save!

      redirect_to db_work_trailers_path(@work), notice: t("messages._common.created")
    end

    def edit
      @trailer = Trailer.find(params[:id])
      authorize @trailer, :edit?
      @work = @trailer.work
    end

    def update
      @trailer = Trailer.find(params[:id])
      authorize @trailer, :update?

      @trailer.attributes = trailer_params
      @trailer.user = current_user

      return render(:edit) unless @trailer.valid?
      @trailer.save_and_create_activity!

      redirect_to db_work_trailers_path(@trailer.work), notice: t("messages._common.updated")
    end

    def hide
      @trailer = Trailer.find(params[:id])
      authorize @trailer, :hide?

      @trailer.soft_delete

      flash[:notice] = t("resources.trailer.unpublished")
      redirect_back fallback_location: db_work_trailers_path(@trailer.work)
    end

    def destroy
      @trailer = Trailer.find(params[:id])
      @trailer.destroy

      flash[:notice] = t("resources.trailer.deleted")
      redirect_back fallback_location: db_work_trailers_path(@trailer.work)
    end

    def activities
      @trailer = Trailer.find(params[:id])
      @activities = @trailer.db_activities.order(id: :desc)
      @comment = @trailer.db_comments.new
    end

    private

    def trailer_rows_form_params
      params.require(:db_trailer_rows_form).permit(:rows)
    end

    def trailer_params
      params.require(:trailer).permit(:title, :url, :sort_number)
    end
  end
end
