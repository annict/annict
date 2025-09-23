# typed: false
# frozen_string_literal: true

RSpec.describe "PATCH /fragment/works/:work_id/note", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトすること" do
    work = FactoryBot.create(:work)

    patch fragment_note_path(work), params: {
      forms_note_form: {body: "新しいメモ"}
    }

    expect(response).to redirect_to(new_user_session_path)
  end

  it "ログインしているとき、メモを更新できること" do
    user = FactoryBot.create(:user)
    work = FactoryBot.create(:work)
    library_entry = FactoryBot.create(:library_entry, user:, work:, note: "古いメモ")

    login_as(user, scope: :user)

    patch fragment_note_path(work), params: {
      forms_note_form: {body: "新しいメモ"}
    }

    expect(response).to redirect_to(fragment_edit_note_path(work))
    expect(flash[:notice]).to eq(I18n.t("messages._common.updated"))

    library_entry.reload
    expect(library_entry.note).to eq("新しいメモ")
  end

  it "ログインしているが、ライブラリエントリが存在しないとき、404エラーになること" do
    user = FactoryBot.create(:user)
    work = FactoryBot.create(:work)

    login_as(user, scope: :user)

    patch fragment_note_path(work), params: {
      forms_note_form: {body: "新しいメモ"}
    }

    expect(response.status).to eq(404)
  end

  it "ログインしているが、作品が削除済みのとき、404エラーになること" do
    user = FactoryBot.create(:user)
    work = FactoryBot.create(:work, deleted_at: Time.current)
    FactoryBot.create(:library_entry, user:, work:, note: "古いメモ")

    login_as(user, scope: :user)

    patch fragment_note_path(work), params: {
      forms_note_form: {body: "新しいメモ"}
    }

    expect(response.status).to eq(404)
  end

  it "ログインしていて、フォームが無効なとき、編集画面を再表示すること" do
    user = FactoryBot.create(:user)
    work = FactoryBot.create(:work)
    FactoryBot.create(:library_entry, user:, work:, note: "古いメモ")

    login_as(user, scope: :user)

    # bodyが長すぎる場合を想定（最大文字数: 1,048,596）
    patch fragment_note_path(work), params: {
      forms_note_form: {body: "a" * 1_048_597}
    }

    expect(response).to have_http_status(:unprocessable_entity)
    # フォームが再表示されることを確認
    expect(response.body).to include("forms_note_form")
  end

  it "ログインしているとき、メモを空文字で更新できること" do
    user = FactoryBot.create(:user)
    work = FactoryBot.create(:work)
    library_entry = FactoryBot.create(:library_entry, user:, work:, note: "既存のメモ")

    login_as(user, scope: :user)

    patch fragment_note_path(work), params: {
      forms_note_form: {body: ""}
    }

    expect(response).to redirect_to(fragment_edit_note_path(work))
    expect(flash[:notice]).to eq(I18n.t("messages._common.updated"))

    library_entry.reload
    expect(library_entry.note).to eq("")
  end

  it "ログインしているとき、メモの前後の空白を削除して更新すること" do
    user = FactoryBot.create(:user)
    work = FactoryBot.create(:work)
    library_entry = FactoryBot.create(:library_entry, user:, work:, note: "古いメモ")

    login_as(user, scope: :user)

    patch fragment_note_path(work), params: {
      forms_note_form: {body: "  前後に空白があるメモ  "}
    }

    expect(response).to redirect_to(fragment_edit_note_path(work))

    library_entry.reload
    expect(library_entry.note).to eq("前後に空白があるメモ")
  end

  it "存在しない作品IDで404エラーになること" do
    user = FactoryBot.create(:user)

    login_as(user, scope: :user)

    patch fragment_note_path(999999), params: {
      forms_note_form: {body: "新しいメモ"}
    }

    expect(response.status).to eq(404)
  end
end
