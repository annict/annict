# typed: false
# frozen_string_literal: true

RSpec.describe "GET /fragment/works/:work_id/note/edit", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトすること" do
    work = FactoryBot.create(:work)

    get fragment_edit_note_path(work)

    expect(response).to redirect_to(new_user_session_path)
  end

  it "ログインしているとき、編集フォームが表示されること" do
    user = FactoryBot.create(:user)
    work = FactoryBot.create(:work)
    FactoryBot.create(:library_entry, user:, work:, note: "既存のメモ")

    login_as(user, scope: :user)

    get fragment_edit_note_path(work)

    expect(response).to have_http_status(:ok)
    expect(response.body).to include("既存のメモ")
  end

  it "ログインしているが、ライブラリエントリが存在しないとき、404エラーになること" do
    user = FactoryBot.create(:user)
    work = FactoryBot.create(:work)

    login_as(user, scope: :user)

    get fragment_edit_note_path(work)

    expect(response.status).to eq(404)
  end

  it "ログインしているが、作品が削除済みのとき、404エラーになること" do
    user = FactoryBot.create(:user)
    work = FactoryBot.create(:work, deleted_at: Time.current)
    FactoryBot.create(:library_entry, user:, work:, note: "既存のメモ")

    login_as(user, scope: :user)

    get fragment_edit_note_path(work)

    expect(response.status).to eq(404)
  end

  it "存在しない作品IDで404エラーになること" do
    user = FactoryBot.create(:user)

    login_as(user, scope: :user)

    get fragment_edit_note_path(999999)

    expect(response.status).to eq(404)
  end

  it "ログインしているとき、メモが空の場合も編集フォームが表示されること" do
    user = FactoryBot.create(:user)
    work = FactoryBot.create(:work)
    FactoryBot.create(:library_entry, user:, work:, note: "")

    login_as(user, scope: :user)

    get fragment_edit_note_path(work)

    expect(response).to have_http_status(:ok)
    # フォームにメモが空の状態で表示されることを確認
    expect(response.body).to include('name="forms_note_form[body]"')
  end

  it "ログインしているとき、メモがデフォルト値の場合も編集フォームが表示されること" do
    user = FactoryBot.create(:user)
    work = FactoryBot.create(:work)
    FactoryBot.create(:library_entry, user:, work:)

    login_as(user, scope: :user)

    get fragment_edit_note_path(work)

    expect(response).to have_http_status(:ok)
    # フォームが表示されることを確認
    expect(response.body).to include('name="forms_note_form[body]"')
  end
end
