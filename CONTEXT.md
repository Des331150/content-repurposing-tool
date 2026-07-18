# Content Repurposing Tool

A personal tool that ingests a single long-form video and produces multiple short clips formatted for social media platforms, using AI to identify narrative highlights and clean visual segments.

## Language

**Source Video**:
A single long-form video uploaded by the user. The atomic input to the system.
_Avoid_: Raw footage, master file

**Short Clip**:
A generated output segment extracted from the source video, formatted for a target platform.
_Avoid_: Highlight, cut, snippet

**Narrative Moment**:
A segment of the source video identified by AI transcript analysis as having narrative value (a quotable insight, a story beat, a punchline).
_Avoid_: Highlight, key moment

**Scene Boundary**:
A clean visual cut point detected in the source video between distinct scenes.
_Avoid_: Cut, transition point

**Generation Run**:
A single processing session that produces a set of short clips from one source video.
_Avoid_: Batch, job, session

**Target Platform**:
A social media platform with specific formatting requirements, e.g. TikTok, YouTube Shorts, Instagram Reels.
_Avoid_: Platform, channel

**Pipeline**:
The linear sequence of processing stages: Audio Extraction → Transcription → Narrative Analysis → Scene Detection → Clip Selection → Platform Formatting → Review → Export.
_Avoid_: Workflow, process

**Review Step**:
The stage where the user can accept, reject, or edit AI-suggested clips before final export.
_Avoid_: Approval, moderation

**Platform Formatting**:
The stage where a clip is adapted to a target platform's requirements — aspect ratio, duration limits, caption style.
_Avoid_: Rendering, adaptation
