# frozen_string_literal: true

module Fragment
  class NotesController < Fragment::ApplicationController
    before_action :authenticate_user!

    def edit
      @work = Work.only_kept.find(params[:work_id])
      library_entry = current_user.library_entries.find_by!(work: @work)
      @form = Forms::NoteForm.new(body: library_entry.note)
    end

    def update
      @work = Work.only_kept.find(params[:work_id])
      library_entry = current_user.library_entries.find_by!(work: @work)
      @form = Forms::NoteForm.new(note_form_params)
      @form.library_entry = library_entry

      if @form.invalid?
        return render :edit, status: :unprocessable_entity
      end

      Updaters::NoteUpdater.new(form: @form).call

      flash[:notice] = t "messages._common.updated"
      redirect_to fragment_edit_note_path(@work)
    end

    private

    def note_form_params
      params.required(:forms_note_form).permit(:body)
    end
  end
end
