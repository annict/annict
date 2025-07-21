# typed: false
# frozen_string_literal: true

RSpec.describe "POST /db/works/:work_id/slots", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトすること" do
    program = create(:program)
    work = create(:work)
    form_params = {
      rows: "#{program.id},2020-04-01 0:00"
    }

    post "/db/works/#{work.id}/slots", params: {deprecated_db_slot_rows_form: form_params}

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
    expect(Slot.all.size).to eq(0)
  end

  it "エディター権限を持たないユーザーがログインしているとき、アクセスできないこと" do
    program = create(:program)
    work = create(:work)
    user = create(:registered_user)
    form_params = {
      rows: "#{program.id},2020-04-01 0:00"
    }

    login_as(user, scope: :user)

    post "/db/works/#{work.id}/slots", params: {deprecated_db_slot_rows_form: form_params}

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
    expect(Slot.all.size).to eq(0)
  end

  it "エディター権限を持つユーザーがログインしているとき、スロットを作成できること" do
    program = create(:program)
    work = create(:work)
    user = create(:registered_user, :with_editor_role)
    form_params = {
      rows: "#{program.id},2020-04-01 0:00"
    }

    login_as(user, scope: :user)

    expect(Slot.all.size).to eq(0)

    post "/db/works/#{work.id}/slots", params: {deprecated_db_slot_rows_form: form_params}

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("登録しました")
    expect(Slot.all.size).to eq(1)

    slot = Slot.last
    expect(slot.program_id).to eq(program.id)
    expect(slot.started_at.to_s).to eq(Time.zone.parse("2020-03-31 15:00").to_s)
  end

  it "複数行のスロットデータを一度に作成できること" do
    work = create(:work)
    program1 = create(:program, work:)
    program2 = create(:program, work:)
    user = create(:registered_user, :with_editor_role)
    form_params = {
      rows: "#{program1.id},2020-04-01 0:00\n#{program2.id},2020-04-02 1:30"
    }

    login_as(user, scope: :user)

    expect(Slot.all.size).to eq(0)

    post "/db/works/#{work.id}/slots", params: {deprecated_db_slot_rows_form: form_params}

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("登録しました")
    expect(Slot.all.size).to eq(2)

    slots = Slot.order(:started_at)
    expect(slots[0].program_id).to eq(program1.id)
    expect(slots[0].started_at.to_s).to eq(Time.zone.parse("2020-03-31 15:00").to_s)
    expect(slots[1].program_id).to eq(program2.id)
    expect(slots[1].started_at.to_s).to eq(Time.zone.parse("2020-04-01 16:30").to_s)
  end

  it "存在しないプログラムIDを指定したとき、エラーになること" do
    work = create(:work)
    user = create(:registered_user, :with_editor_role)
    form_params = {
      rows: "non-existent-id,2020-04-01 0:00"
    }

    login_as(user, scope: :user)

    expect do
      post "/db/works/#{work.id}/slots", params: {deprecated_db_slot_rows_form: form_params}
    end.to raise_error(ActionView::Template::Error)

    expect(Slot.all.size).to eq(0)
  end

  it "不正な日時フォーマットを指定したとき、エラーになること" do
    program = create(:program)
    work = create(:work)
    user = create(:registered_user, :with_editor_role)
    form_params = {
      rows: "#{program.id},invalid-date-format"
    }

    login_as(user, scope: :user)

    expect do
      post "/db/works/#{work.id}/slots", params: {deprecated_db_slot_rows_form: form_params}
    end.to raise_error(ActionView::Template::Error)

    expect(Slot.all.size).to eq(0)
  end

  it "空のrowsパラメータを送信したとき、エラーになること" do
    work = create(:work)
    user = create(:registered_user, :with_editor_role)
    form_params = {
      rows: ""
    }

    login_as(user, scope: :user)

    expect do
      post "/db/works/#{work.id}/slots", params: {deprecated_db_slot_rows_form: form_params}
    end.to raise_error(ActionView::Template::Error)

    expect(Slot.all.size).to eq(0)
  end

  it "存在しない作品IDを指定したとき、404エラーが返ること" do
    user = create(:registered_user, :with_editor_role)

    login_as(user, scope: :user)

    expect do
      post "/db/works/non-existent-id/slots", params: {deprecated_db_slot_rows_form: {rows: ""}}
    end.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "削除済みの作品を指定したとき、404エラーが返ること" do
    work = create(:work, :deleted)
    user = create(:registered_user, :with_editor_role)

    login_as(user, scope: :user)

    expect do
      post "/db/works/#{work.id}/slots", params: {deprecated_db_slot_rows_form: {rows: ""}}
    end.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "異なる作品のプログラムIDを指定したとき、スロットが作成されること" do
    work1 = create(:work)
    work2 = create(:work)
    program = create(:program, work: work2)
    user = create(:registered_user, :with_editor_role)
    form_params = {
      rows: "#{program.id},2020-04-01 0:00"
    }

    login_as(user, scope: :user)

    post "/db/works/#{work1.id}/slots", params: {deprecated_db_slot_rows_form: form_params}

    # フォームのバリデーションでは別作品のプログラムもチェックされない
    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("登録しました")
    expect(Slot.all.size).to eq(1)
  end

  it "同じ日時・プログラムの重複したスロットは作成されること" do
    work = create(:work)
    program = create(:program, work:)
    user = create(:registered_user, :with_editor_role)

    # 最初のスロットを作成
    create(:slot, program:, started_at: Time.zone.parse("2020-03-31 15:00"))

    form_params = {
      rows: "#{program.id},2020-04-01 0:00"
    }

    login_as(user, scope: :user)

    post "/db/works/#{work.id}/slots", params: {deprecated_db_slot_rows_form: form_params}

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("登録しました")
    # 重複チェックはないので、同じ日時でも作成される
    expect(Slot.all.size).to eq(2)
  end
end
