"""Synthetic full exam DOM; no private source content or network dependency."""

import hashlib


def element(tag="div", cls="", children=None, **attrs):
    return {"tag": tag, "class": cls, "children": children or [], **attrs}


def paragraph(value):
    return element("p", children=[{"text": value}])


def transcript(lines):
    return element(
        cls="context-content context-transcript",
        children=[element(cls="collapse", children=[paragraph(line) for line in lines])],
    )


def full_export():
    parts = []
    ranges = ((1, 6, 1), (7, 31, 1), (32, 70, 3), (71, 100, 3), (101, 130, 1), (131, 146, 4), (147, 175, 0))
    groups = [list(range(lo, hi + 1, step)) for lo, hi, step in ranges[:-1]]
    groups.append([147, 149, 152, 155, 158, 161, 164, 167, 170, 173, 176, 181, 186, 191, 196])
    for part, firsts in enumerate(groups, 1):
        wrappers = []
        for index, first in enumerate(firsts):
            width = (
                1
                if part in {1, 2, 5}
                else 3
                if part in {3, 4}
                else 4
                if part == 6
                else (firsts[index + 1] if index + 1 < len(firsts) else 201) - first
            )
            questions = []
            for number in range(first, first + width):
                choices = []
                for key in "ABC" if part == 2 else "ABCD":
                    choices.append(
                        element(
                            cls="form-check",
                            children=[
                                element("input", "form-check-input", value=key),
                                element(
                                    "label",
                                    children=[{"text": f"{key}." + (f" Answer {key}" if part >= 3 else "")}],
                                ),
                            ],
                        )
                    )
                questions.append(
                    element(
                        cls="question-item-wrapper",
                        children=[
                            element(
                                cls="question-wrapper",
                                children=[
                                    element(cls="question-number", children=[{"text": str(number)}]),
                                    element(cls="question-text", children=[paragraph(f"Question {number}?")]),
                                    element(cls="question-answers", children=choices),
                                    element(cls="text-success", children=[{"text": "Đáp án đúng:A"}]),
                                ],
                            ),
                            element(
                                cls="question-explanation-wrapper",
                                children=[
                                    element(cls="collapse", children=[paragraph("Source explanation")])
                                ],
                            ),
                        ],
                    )
                )
            context = []
            if part <= 2:
                lines = ([] if part == 1 else ["Who is speaking?"]) + [
                    f"({key}) Answer {key}." for key in ("ABC" if part == 2 else "ABCD")
                ]
                context.append(transcript(lines))
            elif part <= 4:
                context.append(transcript(["M1: Source utterance.", "Dịch nghĩa:", "M1: Bản dịch."]))
            if part == 1:
                context.append(
                    element(cls="context-content context-image", children=[element("img", asset="photo")])
                )
            if part in {6, 7}:
                passage_count = 1 if part == 6 or first <= 175 else 2 if first <= 185 else 3
                passages = []
                for p in range(passage_count):
                    if p:
                        passages.append(paragraph("----------------------"))
                    passage = [
                        paragraph(
                            "Source passage "
                            + (
                                " ".join(f"before --({q})-- after" for q in range(first, first + width))
                                if part == 6
                                else str(p + 1)
                            )
                        )
                    ]
                    if part == 7:
                        passage.append(
                            element(
                                "table",
                                children=[
                                    element("tr", children=[element("td", children=[{"text": "Cell"}])])
                                ],
                            )
                        )
                    passages.append(element(children=passage))
                context.append(element(cls="context-content", children=[element(children=passages)]))
            if width == 1:
                wrapper = questions[0]
                wrapper["children"].insert(0, element(cls="context-wrapper", children=context))
            else:
                wrapper = element(
                    cls="question-group-wrapper",
                    children=[
                        element(cls="context-wrapper", children=context),
                        element(cls="questions-wrapper", children=questions),
                    ],
                )
            wrappers.append(wrapper)
        parts.append(element(cls="test-questions-wrapper", children=wrappers))
    return {
        "schema_version": 1,
        "page_kind": "exam",
        "source": {"url": "https://study4.com/tests/9999/review/", "title": "Synthetic TOEIC"},
        "content": [element(cls="test-content", children=parts)],
        "assets": [
            {
                "id": "audio",
                "kind": "audio",
                "status": "downloaded",
                "sourceId": "audio.mp3",
                "localPath": "cache/audio_full.mp3",
                "sha256": hashlib.sha256(b"mock_audio_bytes").hexdigest(),
                "url": "https://media.example/audio.mp3",
            },
            {
                "id": "photo",
                "kind": "image",
                "status": "downloaded",
                "sourceId": "photo.png",
                "sha256": "b" * 64,
                "url": "https://media.example/photo.png",
            },
        ],
    }


def listening_segments(data):
    import sys
    from pathlib import Path

    sys.path.insert(0, str(Path(__file__).resolve().parents[3] / "scripts"))
    from align_study4_audio import get_listening_items_from_dom

    segments = []
    for index, item in enumerate(get_listening_items_from_dom(data["content"])):
        number = item["questions"][0]
        start = 5 + index * 20
        segments.extend(
            [
                {
                    "start": start,
                    "end": start + 2,
                    "text": f"{'Number' if number <= 31 else 'Questions'} {number}.",
                },
                {"start": start + 3, "end": start + 15, "text": "Source speech."},
            ]
        )
    return segments
